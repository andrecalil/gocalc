import type { ReactNode } from "react";
import { render, type RenderOptions } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toast } from "@heroui/react";

/**
 * Minimal render wrapper for component tests — sets up QueryClient and the
 * HeroUI Toast.Provider. A fresh QueryClient per render keeps tests isolated
 * (no cache bleed between them).
 */
export function renderWithProviders(ui: ReactNode, opts?: RenderOptions) {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <Toast.Provider />
      {ui}
    </QueryClientProvider>,
    opts,
  );
}
