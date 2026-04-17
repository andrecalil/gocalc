import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

export interface HistoryEntry {
  expression: string;
  result: string;
}

export interface HistoryState {
  history: HistoryEntry[];
  add: (entry: HistoryEntry) => void;
  clear: () => void;
}

const MAX_HISTORY = 5;

/**
 * Persisted list of the most recent calculations (cap: MAX_HISTORY).
 *
 * Insertion rules:
 *  - If the new expression already exists anywhere in the list, REMOVE that
 *    existing entry and prepend the new one. Recalling a history entry and
 *    resubmitting moves it to the top instead of leaving a stale duplicate.
 *  - Cap at MAX_HISTORY (newest first, oldest drops off).
 *
 * Persisted under a versioned localStorage key so the schema can evolve
 * without colliding with old data.
 */
export const useHistoryStore = create<HistoryState>()(
  persist(
    (set) => ({
      history: [],
      add: (entry) =>
        set((state) => {
          // Drop any existing entry with the same expression; the new one
          // (which may have a refreshed result, though for pure arithmetic
          // it will match) replaces it at the top.
          const withoutDup = state.history.filter(
            (e) => e.expression !== entry.expression,
          );
          return {
            history: [entry, ...withoutDup].slice(0, MAX_HISTORY),
          };
        }),
      clear: () => set({ history: [] }),
    }),
    {
      name: "calculator:history:v1",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
