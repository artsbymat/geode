document.addEventListener("DOMContentLoaded", () => {
  const callouts = document.querySelectorAll(".callout.is-collapsible");

  callouts.forEach((callout) => {
    const title = callout.querySelector(".callout-title");
    if (title) {
      title.addEventListener("click", () => {
        callout.classList.toggle("is-collapsed");
      });
    }
  });
});
