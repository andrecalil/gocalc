import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App } from "./App";
import "./styles/globals.css";

// Single QueryClient for the app. Mutations don't retry by default here;
// calculation errors aren't transient, so retrying would just spam the toast.
const queryClient = new QueryClient({
  defaultOptions: {
    mutations: { retry: false },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </StrictMode>,
);
