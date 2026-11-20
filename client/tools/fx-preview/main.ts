// The fx preview harness's entry (#4868). A dev tool: `npm run fx-preview`
// serves it with its own Vite config, and the app's build never sees it.
import { mount } from "svelte";
import "../../src/app.css";
import Harness from "./Harness.svelte";
import { installClock } from "./clock";

// Before anything schedules a timer, so Pause can hold them all.
installClock();

const target = document.getElementById("app");
if (!target) throw new Error("missing #app element");
mount(Harness, { target });
