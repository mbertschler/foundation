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

console.log("Hello from Foundation!");
