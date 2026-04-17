import { useCalcStore } from "../stores/calcStore";

/**
 * The two stacked read-only panels at the top of the calculator:
 *   - expression (what the user is composing)
 *   - result (the latest successful calculation)
 *
 * Both are monospace with tabular figures so digits line up. Neither is
 * editable via this component — input lives in the Keypad + keyboard handler.
 */
export function Display() {
  const expression = useCalcStore((s) => s.expression);
  const result = useCalcStore((s) => s.result);

  return (
    <div
      className="calc-display flex flex-col gap-1 rounded-lg bg-[var(--surface-secondary)] p-4 text-right"
      aria-live="polite"
    >
      <div
        className="min-h-[1.5rem] text-sm text-[var(--muted)] truncate-expr"
        data-testid="display-expression"
        title={expression}
      >
        {expression || "\u00A0"}
      </div>
      <div
        className="min-h-[2.5rem] text-3xl font-medium text-[var(--foreground)] truncate-expr"
        data-testid="display-result"
      >
        {result ?? "0"}
      </div>
    </div>
  );
}
