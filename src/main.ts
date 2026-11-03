import { mount } from "svelte";
import "./lib/theme.svelte";
import "./styles/base.css";
import "./styles/components.css";
import "./styles/login.css";
import "./styles/dashboard.css";
import "./styles/danmaku.css";
import "./styles/side.css";
import App from "./App.svelte";

mount(App, { target: document.querySelector<HTMLElement>("#app")! });
