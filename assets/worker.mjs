// Minimal source fixture; this bundle has not been deployed.
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname === "/health") return new Response("ok");
    return new Response("UICL Worker example", {headers: {"Content-Type": "text/plain; charset=utf-8"}});
  }
};
