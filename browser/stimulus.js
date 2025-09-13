import { Application } from "@hotwired/stimulus";

import ToastButton from "./controllers/toast_button";

export function setupStimulus() {
  window.Stimulus = Application.start();
  Stimulus.register("toast-button", ToastButton);
}

export default {
  setupStimulus,
};
