import { Button } from "@heroui/react";
import { useCalcStore } from "../stores/calcStore";
import { useCalculateMutation } from "../hooks/useCalculateMutation";
import { JSX } from "react/jsx-dev-runtime";

type KeyRole =
  | "digit"
  | "operator"
  | "equals"
  | "clear"
  | "function"
  | "spacer"
  | "backspace";

interface Key {
  label: string | JSX.Element; // What to show on the button
  // What to append to the expression. For functions like sqrt we append
  // `sqrt(` so the user can type the argument and close the paren.
  insert?: string;
  role: KeyRole;
  action?: "equals" | "clear" | "backspace";
}

// The layout mirrors a traditional calculator keypad. The top row holds the
// utility keys (Clear, Backspace, parens, sqrt) and is 5-wide; everything
// below is a standard 4-wide grid.
const KEYS: Key[][] = [
  [
    { label: "Clear", role: "clear", action: "clear" },
    { label: <>&#x232B;</>, role: "backspace", action: "backspace" },
    { label: "(", insert: "(", role: "operator" },
    { label: ")", insert: ")", role: "operator" },
  ],
  [
    { label: "7", insert: "7", role: "digit" },
    { label: "8", insert: "8", role: "digit" },
    { label: "9", insert: "9", role: "digit" },
    { label: "/", insert: "/", role: "operator" },
  ],
  [
    { label: "4", insert: "4", role: "digit" },
    { label: "5", insert: "5", role: "digit" },
    { label: "6", insert: "6", role: "digit" },
    { label: "*", insert: "*", role: "operator" },
  ],
  [
    { label: "1", insert: "1", role: "digit" },
    { label: "2", insert: "2", role: "digit" },
    { label: "3", insert: "3", role: "digit" },
    { label: "-", insert: "-", role: "operator" },
  ],
  [
    { label: "0", insert: "0", role: "digit" },
    { label: ".", insert: ".", role: "digit" },
    { label: "^", insert: "^", role: "operator" },
    { label: "+", insert: "+", role: "operator" },
  ],
  [
    { label: "%", insert: "%", role: "operator" },
    { label: <>&radic;</>, insert: "sqrt(", role: "function" },
    { label: "", insert: "", role: "spacer" },
    { label: "=", role: "equals", action: "equals" },
  ],
];

/** Button variant per key role. See HeroUI v3 button variants. */
function variantFor(role: KeyRole): "primary" | "secondary" | "tertiary" | "outline" | "danger-soft" {
  switch (role) {
    case "equals":
      return "primary"; // gold — the accent cascades here
    case "clear":
      return "danger-soft";
    case "backspace":
    case "operator":
    case "function":
      return "secondary";
    case "digit":
    default:
      return "tertiary";
  }
}

export function Keypad() {
  const append = useCalcStore((s) => s.appendToExpression);
  const clear = useCalcStore((s) => s.clear);
  const backspace = useCalcStore((s) => s.backspace);
  const expression = useCalcStore((s) => s.expression);
  const mutation = useCalculateMutation();

  function handleKey(key: Key) {
    if (key.action === "clear") {
      clear();
      return;
    }
    if (key.action === "backspace") {
      backspace();
      return;
    }
    if (key.action === "equals") {
      // Defense in depth: the API also rejects empty strings, but we skip
      // the round trip.
      if (expression.trim() === "") return;
      mutation.mutate(expression);
      return;
    }
    if (key.insert != null) {
      append(key.insert);
    }
  }

  return (
    <div className="flex flex-col gap-2" data-testid="keypad">
      {KEYS.map((row, i) => {
        return (
          <div key={i} className="grid gap-2 grid-cols-4">
          {row.map((key) => (
            key.role === "spacer" ? <>&nbsp;</> :
            <Button
              key={`calc-key-${i}`}
              variant={variantFor(key.role)}
              className={"w-full p-8 text-lg"}
              size="lg"
              onPress={() => handleKey(key)}
              isDisabled={key.action === "equals" && mutation.isPending}
              data-testid={`key-${key.label}`}
            >
              {key.label}
            </Button>
            ))}
          </div>
        );
      })}
    </div>
  );
}
