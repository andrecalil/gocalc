import { beforeEach, describe, expect, it } from "vitest";
import { useCalcStore } from "./calcStore";

describe("calcStore", () => {
  beforeEach(() => {
    useCalcStore.setState({ expression: "", result: null });
  });

  it("appends characters to the expression", () => {
    useCalcStore.getState().appendToExpression("1");
    useCalcStore.getState().appendToExpression("+2");
    expect(useCalcStore.getState().expression).toBe("1+2");
  });

  it("setExpression replaces wholesale", () => {
    useCalcStore.getState().setExpression("3*4");
    expect(useCalcStore.getState().expression).toBe("3*4");
  });

  it("backspace removes the last character", () => {
    useCalcStore.setState({ expression: "123" });
    useCalcStore.getState().backspace();
    expect(useCalcStore.getState().expression).toBe("12");
  });

  it("backspace on empty is a no-op", () => {
    useCalcStore.getState().backspace();
    expect(useCalcStore.getState().expression).toBe("");
  });

  it("clear wipes both expression and result", () => {
    useCalcStore.setState({ expression: "1+1", result: "2" });
    useCalcStore.getState().clear();
    expect(useCalcStore.getState().expression).toBe("");
    expect(useCalcStore.getState().result).toBeNull();
  });

  it("setResult updates only result", () => {
    useCalcStore.setState({ expression: "1+1" });
    useCalcStore.getState().setResult("2");
    expect(useCalcStore.getState().result).toBe("2");
    expect(useCalcStore.getState().expression).toBe("1+1");
  });
});
