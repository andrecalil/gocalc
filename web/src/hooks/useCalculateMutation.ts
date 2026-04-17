import { useMutation } from "@tanstack/react-query";
import { toast } from "@heroui/react";
import { calculate, CalculateApiError } from "../api/calculate";
import { useCalcStore } from "../stores/calcStore";
import { useHistoryStore } from "../stores/historyStore";

/**
 * Wraps the calculate API call in a TanStack Query mutation and plugs it into
 * the app stores. All error paths — parse failures, 4xx/5xx, network errors —
 * funnel through onError into a toast, so callers don't need their own
 * try/catch.
 */
export function useCalculateMutation() {
  const setResult = useCalcStore((s) => s.setResult);
  const addHistory = useHistoryStore((s) => s.add);

  return useMutation<string, CalculateApiError, string>({
    mutationFn: calculate,
    onSuccess: (result, expression) => {
      setResult(result);
      addHistory({ expression, result });
    },
    onError: (err) => {
      toast.danger(err.message);
    },
    // Calculation errors aren't transient — don't waste a round trip
    // retrying.
    retry: false,
  });
}
