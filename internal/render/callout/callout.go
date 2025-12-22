package callout

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// AST Nodes

var KindCallout = ast.NewNodeKind("Callout")

type Callout struct {
	ast.BaseBlock
}

func NewCallout() *Callout {
	return &Callout{
		BaseBlock: ast.BaseBlock{},
	}
}

func (c *Callout) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, map[string]string{}, nil)
}

func (c *Callout) Kind() ast.NodeKind {
	return KindCallout
}

var KindCalloutTitle = ast.NewNodeKind("CalloutTitle")

type CalloutTitle struct {
	ast.BaseBlock
	DefaultTitle []byte
}

func NewCalloutTitle() *CalloutTitle {
	return &CalloutTitle{BaseBlock: ast.BaseBlock{}}
}

func (c *CalloutTitle) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, map[string]string{}, nil)
}

func (c *CalloutTitle) Kind() ast.NodeKind {
	return KindCalloutTitle
}

var KindCalloutContent = ast.NewNodeKind("CalloutContent")

type CalloutContent struct {
	ast.BaseBlock
}

func NewCalloutContent() *CalloutContent {
	return &CalloutContent{BaseBlock: ast.BaseBlock{}}
}

func (c *CalloutContent) Dump(source []byte, level int) {
	ast.DumpHelper(c, source, level, map[string]string{}, nil)
}

func (c *CalloutContent) Kind() ast.NodeKind {
	return KindCalloutContent
}

// Parser

var calloutRegex = regexp.MustCompile(`^\[!([a-zA-Z0-9-]+)\]([+-])?(?:[ \t]+(.*))?$`)

type calloutParser struct {
}

func NewCalloutParser() parser.BlockParser {
	return &calloutParser{}
}

func (b *calloutParser) Trigger() []byte {
	return []byte{'>'}
}

func (b *calloutParser) process(reader text.Reader) bool {
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 || pos >= len(line) || line[pos] != '>' {
		return false
	}
	pos++
	if pos >= len(line) || line[pos] == '\n' {
		reader.Advance(pos)
		return true
	}
	reader.Advance(pos)
	if line[pos] == ' ' || line[pos] == '\t' {
		padding := 0
		if line[pos] == '\t' {
			padding = util.TabWidth(reader.LineOffset()) - 1
		}
		reader.AdvanceAndSetPadding(1, padding)
	}
	return true
}

func (b *calloutParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 {
		return nil, parser.NoChildren
	}
	if pos >= len(line) || line[pos] != '>' {
		return nil, parser.NoChildren
	}

	contentPos := pos + 1
	if contentPos < len(line) && line[contentPos] == ' ' {
		contentPos++
	}

	content := line[contentPos:]

	trimmedContent := bytes.TrimRight(content, "\r\n")

	// Check for [!TYPE]
	matches := calloutRegex.FindSubmatch(trimmedContent)
	if matches == nil {
		return nil, parser.NoChildren
	}

	cType := strings.ToLower(string(matches[1]))
	fold := string(matches[2])
	titleRaw := matches[3]

	node := NewCallout()
	node.SetAttributeString("callout-type", cType)

	if fold == "+" {
		node.SetAttributeString("callout-collapsible", "true")
		node.SetAttributeString("callout-collapsed", "false")
	} else if fold == "-" {
		node.SetAttributeString("callout-collapsible", "true")
		node.SetAttributeString("callout-collapsed", "true")
	} else {
		node.SetAttributeString("callout-collapsible", "false")
		node.SetAttributeString("callout-collapsed", "false")
	}

	if !b.process(reader) {
		return nil, parser.NoChildren
	}
	headerLine, headerRemainder := reader.PeekLine()
	nl := 0
	if len(headerLine) > 0 && headerLine[len(headerLine)-1] == '\n' {
		nl = 1
		if len(headerLine) > 1 && headerLine[len(headerLine)-2] == '\r' {
			nl = 2
		}
	}
	if headerRemainder.Len() > nl {
		reader.Advance(headerRemainder.Len() - nl)
	}

	titleNode := NewCalloutTitle()

	if len(titleRaw) > 0 {
		para := ast.NewParagraph()

		matchLocs := calloutRegex.FindSubmatchIndex(trimmedContent)

		if matchLocs[6] != -1 {
			startInContent := matchLocs[6]
			endInContent := matchLocs[7]

			absoluteStart := segment.Start + contentPos + startInContent
			absoluteEnd := segment.Start + contentPos + endInContent

			para.Lines().Append(text.NewSegment(absoluteStart, absoluteEnd))
			titleNode.AppendChild(titleNode, para)
		}
	} else {
		titleText := []byte(strings.Title(cType))
		titleNode.DefaultTitle = titleText
	}

	node.AppendChild(node, titleNode)

	return node, parser.HasChildren
}

func (b *calloutParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if b.process(reader) {
		return parser.Continue | parser.HasChildren
	}
	return parser.Close
}

func (b *calloutParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	callout := node.(*Callout)

	contentNode := NewCalloutContent()

	titleNode := callout.FirstChild()
	if titleNode != nil && titleNode.Kind() == KindCalloutTitle {
		for child := titleNode.NextSibling(); child != nil; {
			next := child.NextSibling()
			callout.RemoveChild(callout, child)
			contentNode.AppendChild(contentNode, child)
			child = next
		}

		callout.AppendChild(callout, contentNode)
	}
}

func (b *calloutParser) CanInterruptParagraph() bool {
	return true
}

func (b *calloutParser) CanAcceptIndentedLine() bool {
	return false
}

// Renderer

type CalloutRenderer struct {
}

func NewCalloutRenderer() renderer.NodeRenderer {
	return &CalloutRenderer{}
}

func (r *CalloutRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCallout, r.renderCallout)
	reg.Register(KindCalloutTitle, r.renderCalloutTitle)
	reg.Register(KindCalloutContent, r.renderCalloutContent)
}

func (r *CalloutRenderer) renderCallout(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	ctx := node.(*Callout)
	if entering {
		cTypeVal, _ := ctx.AttributeString("callout-type")
		cType := ""
		if cTypeVal != nil {
			if s, ok := cTypeVal.(string); ok {
				cType = strings.ToLower(s)
			}
		}

		collapsibleVal, _ := ctx.AttributeString("callout-collapsible")
		collapsible := "false"
		if collapsibleVal != nil {
			if s, ok := collapsibleVal.(string); ok {
				collapsible = s
			}
		}

		collapsedVal, _ := ctx.AttributeString("callout-collapsed")
		collapsed := "false"
		if collapsedVal != nil {
			if s, ok := collapsedVal.(string); ok {
				collapsed = s
			}
		}

		isCollapsible := collapsible == "true"
		isCollapsed := collapsed == "true"

		var class strings.Builder
		class.WriteString("callout ")
		class.WriteString(cType)
		if isCollapsible {
			class.WriteString(" is-collapsible")
		}
		if isCollapsed {
			class.WriteString(" is-collapsed")
		}

		w.WriteString("<blockquote class=\"")
		w.WriteString(class.String())
		w.WriteString("\" data-callout=\"")
		w.WriteString(cType)
		w.WriteString("\"")

		if isCollapsible {
			w.WriteString(" data-callout-fold")
		}

		w.WriteString(">\n")
	} else {
		w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (r *CalloutRenderer) renderCalloutTitle(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	parent := node.Parent()
	if parent == nil || parent.Kind() != KindCallout {
		return ast.WalkContinue, nil
	}

	ct := node.(*CalloutTitle)

	if entering {
		w.WriteString("<div class=\"callout-title\">\n")
		w.WriteString("<div class=\"callout-icon\"></div>\n")
		w.WriteString("<div class=\"callout-title-inner\">")

		if ct.DefaultTitle != nil {
			w.WriteString("<p>")
			w.Write(ct.DefaultTitle)
			w.WriteString("</p>")
		}
	} else {
		w.WriteString("</div>\n")

		collapsible, _ := parent.AttributeString("callout-collapsible")
		if collapsible != nil {
			if s, ok := collapsible.(string); ok && s == "true" {
				w.WriteString("<div class=\"fold-callout-icon\"></div>")
			}
		}

		w.WriteString("\n</div>\n")
	}
	return ast.WalkContinue, nil
}

func (r *CalloutRenderer) renderCalloutContent(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<div class=\"callout-content\">\n")
	} else {
		w.WriteString("</div>\n")
	}
	return ast.WalkContinue, nil
}

// Extender

type Extender struct {
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewCalloutParser(), 10),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewCalloutRenderer(), 10),
		),
	)
}
