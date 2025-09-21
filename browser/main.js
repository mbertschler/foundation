// TODO: basecoat really should export the dist/js/basecoat.js file.
// Then we could do:
//     import "basecoat-css/basecoat";
//      import "basecoat-css/toast";
// For now let's just import everything, and let esbuild tree-shake the unused stuff
import "basecoat-css/all";

// import * as Turbo from "@hotwired/turbo"
import "@hotwired/turbo";

import { setupStimulus } from "./stimulus";
setupStimulus();

document.addEventListener("turbo:render", (event) => {
  // basecoat should actually do this via a mutation observer,
  // but there seems to be some bug. Quick fix for now:
  window.basecoat.initAll();
});

// CSRF token management for Turbo Frames
document.addEventListener("turbo:before-fetch-response", (event) => {
  const csrfToken =
    event.detail.fetchResponse.response.headers.get("x-csrf-token");
  if (csrfToken) {
    updateCSRFMetaTag(csrfToken);
  }
});

function updateCSRFMetaTag(token) {
  const metaTag = document.querySelector('meta[name="csrf-token"]');
  if (metaTag && metaTag.content !== token) {
    metaTag.content = token;
  }
}

console.log("Hello from Foundation! 🌍");
