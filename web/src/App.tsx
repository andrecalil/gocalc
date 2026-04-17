import { Toast } from "@heroui/react";
import { Calculator } from "./components/Calculator";
import { History } from "./components/History";
import { ThemeToggle } from "./components/ThemeToggle";

export function App() {
  return (
    <>
      <Toast.Provider placement="top end" />
      <div className="relative min-h-screen bg-background/20 text-foreground">
        <header className="flex justify-end  px-4 py-3">
          <ThemeToggle />
        </header>
        {/*
         * Desktop (≥768px): two columns — calculator left, history right.
         * Mobile: single column, history stacked below. Simple and tactile,
         * no drawer/toggle.
         */}
        <main className="mx-auto grid max-w-4xl grid-cols-1 gap-4 p-4 md:grid-cols-[1fr_16rem]">
          <Calculator />
          <History />
        </main>
        <footer className="border-t border-surface-secondary  px-4 py-3 flex flex-row justify-between text-muted fixed bottom-0 w-full bg-background">
          <h1 className="text-md font-semibold">GoCalc</h1><a className="text-sm" target="_blank" href="https://github.com/andrecalil/gocalc">Source</a>
        </footer>
      </div>
    </>
  );
}
