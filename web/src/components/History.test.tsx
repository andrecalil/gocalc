import { beforeEach, describe, expect, it } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { History } from "./History";
import { useCalcStore } from "../stores/calcStore";
import { useHistoryStore } from "../stores/historyStore";
import { renderWithProviders } from "../test/renderApp";

describe("History", () => {
  beforeEach(() => {
    useCalcStore.setState({ expression: "", result: null });
    useHistoryStore.setState({ history: [] });
  });

  it("shows an empty state when there is no history", () => {
    renderWithProviders(<History />);
    expect(screen.getByText(/no calculations yet/i)).toBeInTheDocument();
  });

  it("renders recent entries newest-first", () => {
    useHistoryStore.setState({
      history: [
        { expression: "2*3", result: "6" },
        { expression: "1+1", result: "2" },
      ],
    });
    renderWithProviders(<History />);
    const items = screen.getAllByTestId("history-entry");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent("2*3");
    expect(items[1]).toHaveTextContent("1+1");
  });

  it("clicking an entry restores both expression and result", async () => {
    const user = userEvent.setup();
    // A stale result from some previous calculation — clicking a history
    // entry must overwrite it with the entry's result, not leave it.
    useCalcStore.setState({ expression: "", result: "99" });
    useHistoryStore.setState({
      history: [{ expression: "(1+1)*4", result: "8" }],
    });
    renderWithProviders(<History />);
    await user.click(screen.getByTestId("history-entry"));
    expect(useCalcStore.getState().expression).toBe("(1+1)*4");
    expect(useCalcStore.getState().result).toBe("8");
  });

  it("clear button empties the history", async () => {
    const user = userEvent.setup();
    useHistoryStore.setState({
      history: [{ expression: "1+1", result: "2" }],
    });
    renderWithProviders(<History />);
    await user.click(screen.getByTestId("history-clear"));
    expect(useHistoryStore.getState().history).toEqual([]);
  });
});
