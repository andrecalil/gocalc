import { create } from "zustand";

/**
 * UI state for the current calculator session: the expression the user is
 * composing and the latest successful result. Neither is persisted — the
 * session resets on refresh. Only the history store (below) survives reloads.
 */
export interface CalcState {
  expression: string;
  result: string | null;
  setExpression: (s: string) => void;
  appendToExpression: (s: string) => void;
  backspace: () => void;
  setResult: (r: string) => void;
  // recall restores BOTH the expression and the result — used when the user
  // clicks a history entry. The result comes from history because the entry
  // represents a previously-computed (expression, result) pair and we don't
  // want to leave the display showing a stale result from a different one.
  recall: (expression: string, result: string) => void;
  clear: () => void;
}

export const useCalcStore = create<CalcState>((set) => ({
  expression: "",
  result: null,
  setExpression: (s) => set({ expression: s }),
  appendToExpression: (s) =>
    set((state) => ({ expression: state.expression + s })),
  backspace: () =>
    set((state) => ({ expression: state.expression.slice(0, -1) })),
  setResult: (r) => set({ result: r }),
  recall: (expression, result) => set({ expression, result }),
  // `Clear` / `Esc` wipes both input and result. This is deliberate — the
  // user is signaling "start over".
  clear: () => set({ expression: "", result: null }),
}));
