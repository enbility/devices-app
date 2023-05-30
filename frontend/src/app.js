import "bulma/css/bulma.css";
import { createApp } from "vue";
import App from "./views/App.vue";
import setupRouter from "./router";

const app = createApp(App)
app.use(setupRouter())
window.app = app.mount("#app")
