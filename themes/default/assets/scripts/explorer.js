document.addEventListener("DOMContentLoaded", () => {
  const explorer = document.querySelector(".file-explorer");
  if (!explorer) return;

  const STORAGE_KEY = "geode:explorer:open";
  const readOpenKeys = () => {
    try {
      const raw = sessionStorage.getItem(STORAGE_KEY);
      if (!raw) return new Set();
      const arr = JSON.parse(raw);
      if (!Array.isArray(arr)) return new Set();
      return new Set(arr.filter((x) => typeof x === "string"));
    } catch {
      return new Set();
    }
  };

  const writeOpenKeys = (set) => {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(Array.from(set)));
    } catch {
      // ignore
    }
  };

  const openKeys = readOpenKeys();
  explorer.querySelectorAll("li[data-node-key]").forEach((li) => {
    const key = li.getAttribute("data-node-key");
    if (key && openKeys.has(key)) li.classList.add("open");
  });

  explorer.addEventListener("click", (e) => {
    if (!(e.target instanceof Element)) return;

    const folder = e.target.closest(".folder");
    if (!folder) return;

    if (!explorer.contains(folder)) return;

    const li = folder.parentElement;
    if (!li) return;

    li.classList.toggle("open");

    const key = li.getAttribute("data-node-key");
    if (!key) return;

    if (li.classList.contains("open")) {
      openKeys.add(key);
    } else {
      openKeys.delete(key);
    }
    writeOpenKeys(openKeys);
  });
});
