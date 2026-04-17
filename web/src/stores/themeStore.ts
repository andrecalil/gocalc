import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

export type ThemeMode = "light" | "dark";
export type ThemeChoice = ThemeMode | "system";

interface ThemeState {
  // `system` means "follow OS preference"; on first load we resolve this to
  // prefers-color-scheme at render time. Once the user toggles explicitly,
  // their choice is stored as "light" or "dark" and honored from then on.
  choice: ThemeChoice;
  setChoice: (c: ThemeChoice) => void;
  toggle: () => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      choice: "system",
      setChoice: (c) => set({ choice: c }),
      toggle: () => {
        // A toggle from "system" picks the opposite of whatever the system
        // currently prefers, so the UI visibly flips.
        const current = resolveChoice(get().choice);
        set({ choice: current === "dark" ? "light" : "dark" });
      },
    }),
    {
      name: "calculator:theme:v1",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);

/** Resolve a stored choice to the concrete theme that should be applied. */
export function resolveChoice(c: ThemeChoice): ThemeMode {
  if (c === "system") {
    if (typeof window !== "undefined" && window.matchMedia) {
      return window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light";
    }
    return "light";
  }
  return c;
}
