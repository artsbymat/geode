const modal = document.getElementById("searchModal");
const openBtn = document.querySelector("div.utilities button.search");

openBtn.addEventListener("click", () => {
  modal.classList.add("is-open");
  setTimeout(() => {
    const input = document.querySelector("#search input");
    if (input) input.focus();
  }, 100);
});

modal.addEventListener("click", (e) => {
  if (e.target === modal || e.target.classList.contains("modal-backdrop")) {
    modal.classList.remove("is-open");
  }
});

window.addEventListener("DOMContentLoaded", (event) => {
  new PagefindUI({
    element: "#search",
    showSubResults: true,
    showImages: false,
    autoFocus: true,
    processResult: function (result) {
      result.url = result.url.replace(".html", "");
      result.raw_url = result.raw_url.replace(".html", "");
      result.sub_results.map((sub_result) => {
        sub_result.url = sub_result.url.replace(".html", "");
      });
      return result;
    },
  });
});

window.addEventListener("keydown", (e) => {
  if (e.key === "Escape" && modal.classList.contains("is-open")) {
    modal.classList.remove("is-open");
  } else if ((e.ctrlKey || e.metaKey) && e.key === "k") {
    e.preventDefault();
    modal.classList.toggle("is-open");
    if (modal.classList.contains("is-open")) {
      setTimeout(() => {
        const input = document.querySelector("#search input");
        if (input) input.focus();
      }, 100);
    }
  }
});
