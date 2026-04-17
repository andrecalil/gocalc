import { useEffect } from "react";
import { Button } from "@heroui/react";
import { useThemeStore, resolveChoice } from "../stores/themeStore";

/**
 * Applies the resolved theme to <html data-theme="..."> and renders a
 * toggle button. When the stored choice is "system", we subscribe to
 * prefers-color-scheme so the UI re-syncs if the OS changes (e.g., macOS
 * auto-dark at sunset).
 */
export function ThemeToggle() {
  const choice = useThemeStore((s) => s.choice);
  const toggle = useThemeStore((s) => s.toggle);

  useEffect(() => {
    const applied = resolveChoice(choice);
    document.documentElement.dataset.theme = applied;
    document.documentElement.classList.toggle("dark", applied === "dark");
    document.documentElement.classList.toggle("light", applied === "light");
  }, [choice]);

  useEffect(() => {
    if (choice !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const listener = () => {
      const applied = resolveChoice("system");
      document.documentElement.dataset.theme = applied;
      document.documentElement.classList.toggle("dark", applied === "dark");
      document.documentElement.classList.toggle("light", applied === "light");
    };
    mq.addEventListener("change", listener);
    return () => mq.removeEventListener("change", listener);
  }, [choice]);

  const applied = resolveChoice(choice);

  return (
    <Button
      size="sm"
      variant="ghost"
      onPress={toggle}
      data-testid="theme-toggle"
      aria-label="Toggle theme"
    >
      {applied === "dark" ? "Light mode" : "Dark mode"}
    </Button>
  );
}
