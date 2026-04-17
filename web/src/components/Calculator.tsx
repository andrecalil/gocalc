import { useEffect } from "react";
import { Display } from "./Display";
import { Keypad } from "./Keypad";
import { useCalcStore } from "../stores/calcStore";
import { useCalculateMutation } from "../hooks/useCalculateMutation";

/**
 * Keyboard bindings:
 *   - digits / operators / parens / `.`  → append to expression
 *   - Enter  → submit (same as `=`)
 *   - Esc    → clear input and result
 *   - Backspace → delete last character (bonus)
 *
 * We listen on window so the calculator captures input without needing a
 * focused element. Events that look like a modifier-based shortcut (cmd-R
 * etc.) are ignored.
 */
const ALLOWED_CHARS = /^[0-9+\-*/^%().]$/;

export function Calculator() {
  const append = useCalcStore((s) => s.appendToExpression);
  const clear = useCalcStore((s) => s.clear);
  const backspace = useCalcStore((s) => s.backspace);
  const expression = useCalcStore((s) => s.expression);
  const mutation = useCalculateMutation();

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.metaKey || e.ctrlKey || e.altKey) return;

      if (e.key === "Enter" || e.key === "=") {
        e.preventDefault();
        if (expression.trim() === "") return;
        mutation.mutate(expression);
        return;
      }
      if (e.key === "Escape") {
        e.preventDefault();
        clear();
        return;
      }
      if (e.key === "Backspace") {
        e.preventDefault();
        backspace();
        return;
      }
      if (e.key.length === 1 && ALLOWED_CHARS.test(e.key)) {
        e.preventDefault();
        append(e.key);
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [append, backspace, clear, expression, mutation]);

  return (
    <div className="flex flex-col gap-3 bg-background p-3 rounded-xl shadow" data-testid="calculator">
      <Display />
      <Keypad />
    </div>
  );
}
