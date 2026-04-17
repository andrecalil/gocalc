import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../test/setup";
import { calculate, CalculateApiError } from "./calculate";

describe("calculate()", () => {
  it("returns the result string on 200", async () => {
    const result = await calculate("1+1");
    expect(result).toBe("2");
  });

  it("throws CalculateApiError with the API message on 4xx", async () => {
    server.use(
      http.post("/api/calculate", () =>
        HttpResponse.json({ error: "invalid expression" }, { status: 400 }),
      ),
    );
    await expect(calculate("1+")).rejects.toMatchObject({
      name: "CalculateApiError",
      message: "invalid expression",
      status: 400,
    });
  });

  it("throws CalculateApiError on 5xx", async () => {
    server.use(
      http.post("/api/calculate", () =>
        HttpResponse.json({ error: "internal server error" }, { status: 500 }),
      ),
    );
    await expect(calculate("x")).rejects.toBeInstanceOf(CalculateApiError);
  });

  it("falls back to a generic message when the error body is malformed", async () => {
    server.use(
      http.post("/api/calculate", () =>
        HttpResponse.text("oops", { status: 400 }),
      ),
    );
    const err = (await calculate("1+1").catch((e) => e)) as CalculateApiError;
    expect(err).toBeInstanceOf(CalculateApiError);
    expect(err.message).toBe("Calculation failed");
  });

  it("throws on network failure with a non-null message", async () => {
    server.use(
      http.post("/api/calculate", () => HttpResponse.error()),
    );
    await expect(calculate("1+1")).rejects.toBeInstanceOf(CalculateApiError);
  });
});
