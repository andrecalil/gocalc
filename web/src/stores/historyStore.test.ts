import { beforeEach, describe, expect, it } from "vitest";
import { useHistoryStore } from "./historyStore";

describe("historyStore", () => {
  beforeEach(() => {
    useHistoryStore.setState({ history: [] });
    localStorage.clear();
  });

  it("adds entries newest-first", () => {
    useHistoryStore.getState().add({ expression: "1+1", result: "2" });
    useHistoryStore.getState().add({ expression: "2*3", result: "6" });
    expect(useHistoryStore.getState().history).toEqual([
      { expression: "2*3", result: "6" },
      { expression: "1+1", result: "2" },
    ]);
  });

  it("caps the list at 5 entries", () => {
    for (let i = 0; i < 7; i++) {
      useHistoryStore.getState().add({ expression: `${i}`, result: `${i}` });
    }
    const h = useHistoryStore.getState().history;
    expect(h).toHaveLength(5);
    // Newest first — 6, 5, 4, 3, 2 — old ones (0, 1) dropped.
    expect(h.map((e) => e.expression)).toEqual(["6", "5", "4", "3", "2"]);
  });

  it("moves existing entries to the top on resubmit instead of duplicating", () => {
    // Seed three distinct entries: newest → oldest is [c, b, a].
    useHistoryStore.getState().add({ expression: "a", result: "1" });
    useHistoryStore.getState().add({ expression: "b", result: "2" });
    useHistoryStore.getState().add({ expression: "c", result: "3" });
    expect(useHistoryStore.getState().history.map((e) => e.expression)).toEqual(
      ["c", "b", "a"],
    );

    // Resubmitting "a" (recall + `=`) should lift it to the top and leave
    // the other entries intact — no duplicate of "a" in the list.
    useHistoryStore.getState().add({ expression: "a", result: "1" });
    expect(useHistoryStore.getState().history.map((e) => e.expression)).toEqual(
      ["a", "c", "b"],
    );
  });

  it("resubmitting the most recent entry is still a no-op to position", () => {
    useHistoryStore.getState().add({ expression: "1+1", result: "2" });
    useHistoryStore.getState().add({ expression: "1+1", result: "2" });
    expect(useHistoryStore.getState().history).toHaveLength(1);
  });

  it("persists to localStorage", () => {
    useHistoryStore.getState().add({ expression: "1+1", result: "2" });
    const raw = localStorage.getItem("calculator:history:v1");
    expect(raw).toBeTruthy();
    expect(raw).toContain("1+1");
  });

  it("clear empties the list", () => {
    useHistoryStore.getState().add({ expression: "1+1", result: "2" });
    useHistoryStore.getState().clear();
    expect(useHistoryStore.getState().history).toEqual([]);
  });
});
