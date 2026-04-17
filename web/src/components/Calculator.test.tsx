import { beforeEach, describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { Calculator } from "./Calculator";
import { useCalcStore } from "../stores/calcStore";
import { useHistoryStore } from "../stores/historyStore";
import { renderWithProviders } from "../test/renderApp";
import { server } from "../test/setup";

function resetStores() {
  useCalcStore.setState({ expression: "", result: null });
  useHistoryStore.setState({ history: [] });
  localStorage.clear();
}

describe("Calculator", () => {
  beforeEach(resetStores);

  it("renders empty initial state", () => {
    renderWithProviders(<Calculator />);
    expect(screen.getByTestId("display-result")).toHaveTextContent("0");
  });

  it("appends digits via keypad clicks", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-1"));
    await user.click(screen.getByTestId("key-+"));
    await user.click(screen.getByTestId("key-1"));
    expect(screen.getByTestId("display-expression")).toHaveTextContent("1+1");
  });

  it("Clear wipes both input and result", async () => {
    const user = userEvent.setup();
    useCalcStore.setState({ expression: "1+1", result: "2" });
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-Clear"));
    expect(useCalcStore.getState().expression).toBe("");
    expect(useCalcStore.getState().result).toBeNull();
  });

  it("= fires the mutation and displays the result + writes history", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-1"));
    await user.click(screen.getByTestId("key-+"));
    await user.click(screen.getByTestId("key-1"));
    await user.click(screen.getByTestId("key-="));
    await waitFor(() =>
      expect(screen.getByTestId("display-result")).toHaveTextContent("2"),
    );
    expect(useHistoryStore.getState().history[0]).toEqual({
      expression: "1+1",
      result: "2",
    });
  });

  it("= on an empty expression does not fire a request", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-="));
    // If a request had fired, result would change away from "0".
    expect(screen.getByTestId("display-result")).toHaveTextContent("0");
  });

  it("Enter key submits", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Calculator />);
    useCalcStore.setState({ expression: "1+1" });
    await user.keyboard("{Enter}");
    await waitFor(() =>
      expect(screen.getByTestId("display-result")).toHaveTextContent("2"),
    );
  });

  it("Escape clears", async () => {
    const user = userEvent.setup();
    useCalcStore.setState({ expression: "1+1", result: "2" });
    renderWithProviders(<Calculator />);
    await user.keyboard("{Escape}");
    expect(useCalcStore.getState().expression).toBe("");
    expect(useCalcStore.getState().result).toBeNull();
  });

  it("Backspace deletes the last char", async () => {
    const user = userEvent.setup();
    useCalcStore.setState({ expression: "123" });
    renderWithProviders(<Calculator />);
    await user.keyboard("{Backspace}");
    expect(useCalcStore.getState().expression).toBe("12");
  });

  it("Backspace keypad button deletes the last char", async () => {
    const user = userEvent.setup();
    useCalcStore.setState({ expression: "42" });
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-⌫"));
    expect(useCalcStore.getState().expression).toBe("4");
  });

  it("typing digits on the keyboard appends to the expression", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Calculator />);
    await user.keyboard("2+3");
    expect(useCalcStore.getState().expression).toBe("2+3");
  });

  it("network errors do not update the display", async () => {
    server.use(
      http.post("/api/calculate", () => HttpResponse.error()),
    );
    const user = userEvent.setup();
    useCalcStore.setState({ expression: "1+1" });
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-="));
    // Give the mutation a tick to resolve without hanging the test.
    await new Promise((r) => setTimeout(r, 30));
    expect(screen.getByTestId("display-result")).toHaveTextContent("0");
    expect(useCalcStore.getState().expression).toBe("1+1"); // input preserved
  });

  it("4xx errors show the API error message (input preserved)", async () => {
    server.use(
      http.post("/api/calculate", () =>
        HttpResponse.json({ error: "invalid expression" }, { status: 400 }),
      ),
    );
    const user = userEvent.setup();
    useCalcStore.setState({ expression: "1+" });
    renderWithProviders(<Calculator />);
    await user.click(screen.getByTestId("key-="));
    await new Promise((r) => setTimeout(r, 30));
    expect(useCalcStore.getState().expression).toBe("1+");
    expect(useCalcStore.getState().result).toBeNull();
  });
});
