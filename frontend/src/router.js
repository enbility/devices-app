import { createRouter, createWebHashHistory } from "vue-router";
import MainView from "./views/Main.vue";

export default function setupRouter() {
  const router = createRouter({
    history: createWebHashHistory(),
    routes: [
      {
        path: "/",
        component: MainView,
        props: true,
      },
    ]
  });
  return router;
}
