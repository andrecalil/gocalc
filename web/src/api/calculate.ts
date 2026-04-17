/**
 * Thin fetch wrapper around POST /api/calculate. The backend is the single
 * source of truth for expression evaluation — this module does no parsing
 * and no validation other than surfacing server errors.
 */

export type CalculateSuccess = { result: string };
export type CalculateError = { error: string };

/**
 * Error thrown by calculate() when the API returns a non-2xx response or the
 * network request itself fails. The `message` is a user-ready string
 * (either the API's `error` field or a fallback for network issues), safe to
 * drop directly into a toast.
 */
export class CalculateApiError extends Error {
  readonly status: number | null;
  constructor(message: string, status: number | null) {
    super(message);
    this.name = "CalculateApiError";
    this.status = status;
  }
}

export async function calculate(expression: string): Promise<string> {
  let res: Response;
  try {
    res = await fetch("/api/calculate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expression }),
    });
  } catch (err) {
    // Network-level failure — fetch never returned a response.
    throw new CalculateApiError(
      err instanceof Error && err.message ? err.message : "Network error",
      null,
    );
  }

  // Try to parse JSON on both success and error paths; if the body is
  // malformed, fall back to a generic message.
  let body: Partial<CalculateSuccess & CalculateError> = {};
  try {
    body = (await res.json()) as Partial<CalculateSuccess & CalculateError>;
  } catch {
    // ignore; handled below
  }

  if (!res.ok) {
    throw new CalculateApiError(body.error ?? "Calculation failed", res.status);
  }
  if (typeof body.result !== "string") {
    throw new CalculateApiError("Malformed response", res.status);
  }
  return body.result;
}
