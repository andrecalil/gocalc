import { Button, Card } from "@heroui/react";
import { useCalcStore } from "../stores/calcStore";
import { useHistoryStore } from "../stores/historyStore";

/**
 * The history panel — a list of up to 5 recent `expression = result` pairs.
 * Clicking an entry replaces the current expression with that expression's
 * text. The result display is NOT touched because the user hasn't submitted
 * yet; they might be editing.
 */
export function History() {
  const history = useHistoryStore((s) => s.history);
  const clear = useHistoryStore((s) => s.clear);
  const recall = useCalcStore((s) => s.recall);

  return (
    <Card className="h-full bg-background rounded-xl shadow p-0">
      <Card.Header className="w-full">
        <Card.Title>
          <div className="flex items-center justify-between border-b border-surface-secondary p-3">
            <h3 className="text-lg font-semibold">History</h3>
            {history.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onPress={clear}
                data-testid="history-clear"
              >
                Clear
              </Button>
            )}
          </div>
        </Card.Title>
      </Card.Header>
      <Card.Content className="p-3">
        {history.length === 0 ? (
          <p className="text-sm text-muted">No calculations yet.</p>
        ) : (
          <ul className="flex flex-col gap-1" data-testid="history-list">
            {history.map((entry, idx) => (
              <li key={idx}>
                <button
                  type="button"
                  className="calc-display w-full rounded px-2 py-1 text-left text-sm hover:bg-surface-secondary truncate-expr"
                  onClick={() => recall(entry.expression, entry.result)}
                  title={`${entry.expression} = ${entry.result}`}
                  data-testid="history-entry"
                >
                  {entry.expression} ={" "}
                  <span className="font-medium">{entry.result}</span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </Card.Content>
    </Card>
  );
}
