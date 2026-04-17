import { http, HttpResponse } from "msw";

/**
 * Default MSW handlers — "happy path" behavior that tests can override per
 * case with server.use(...). The handler parses the body, does a tiny
 * amount of fake math, and returns the canned response shape.
 */
export const handlers = [
  http.post("/api/calculate", async ({ request }) => {
    const body = (await request.json()) as { expression?: string };
    const expr = (body?.expression ?? "").trim();
    if (!expr) {
      return HttpResponse.json(
        { error: "expression is empty" },
        { status: 400 },
      );
    }
    // Extremely naive "eval" just for the tests — the real API does the
    // work. Tests that care about specific backend behavior override this.
    if (expr === "1+1") return HttpResponse.json({ result: "2" });
    if (expr === "(1+1)*4") return HttpResponse.json({ result: "8" });
    return HttpResponse.json({ result: `RESULT(${expr})` });
  }),
];
