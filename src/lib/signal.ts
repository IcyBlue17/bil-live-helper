// A tiny reactive core: signals hold values, effects re-run when the
// signals they read change, and scopes dispose the effects of a view.

type Runner = { run(): void; deps: Set<Set<Runner>> };

let active: Runner | null = null;
let scope: (() => void)[] | null = null;
let batchDepth = 0;
const pending = new Set<Runner>();

export interface Signal<T> {
  (): T;
  set(value: T): void;
  update(fn: (value: T) => T): void;
  peek(): T;
}

export function signal<T>(initial: T): Signal<T> {
  let value = initial;
  const subs = new Set<Runner>();
  const read = (() => {
    if (active) {
      subs.add(active);
      active.deps.add(subs);
    }
    return value;
  }) as Signal<T>;
  read.peek = () => value;
  read.set = (next: T) => {
    if (Object.is(next, value)) return;
    value = next;
    for (const r of [...subs]) pending.add(r);
    if (batchDepth === 0) flush();
  };
  read.update = (fn) => read.set(fn(value));
  return read;
}

function flush() {
  while (pending.size) {
    const runners = [...pending];
    pending.clear();
    for (const r of runners) r.run();
  }
}

/** Groups writes so effects run once, after all of them. */
export function batch(fn: () => void) {
  batchDepth++;
  try {
    fn();
  } finally {
    if (--batchDepth === 0) flush();
  }
}

/** Runs fn now and again whenever a signal it read changes. */
export function effect(fn: () => void | (() => void)): () => void {
  let cleanup: void | (() => void);
  let disposed = false;
  const runner: Runner = {
    deps: new Set(),
    run() {
      if (disposed) return;
      unlink();
      if (cleanup) cleanup();
      const prev = active;
      active = runner;
      try {
        cleanup = fn();
      } finally {
        active = prev;
      }
    },
  };
  const unlink = () => {
    for (const d of runner.deps) d.delete(runner);
    runner.deps.clear();
  };
  const dispose = () => {
    disposed = true;
    unlink();
    pending.delete(runner);
    if (cleanup) cleanup();
  };
  scope?.push(dispose);
  runner.run();
  return dispose;
}

export function computed<T>(fn: () => T): () => T {
  const s = signal<T>(undefined as T);
  effect(() => s.set(fn()));
  return () => s();
}

/** Reads signals without subscribing the running effect. */
export function untrack<T>(fn: () => T): T {
  const prev = active;
  active = null;
  try {
    return fn();
  } finally {
    active = prev;
  }
}

/** Collects the effects and cleanups made by fn; returns its result and a disposer. */
export function root<T>(fn: () => T): [T, () => void] {
  const prev = scope;
  const own: (() => void)[] = [];
  scope = own;
  const prevActive = active;
  active = null;
  try {
    const value = fn();
    return [value, () => own.splice(0).reverse().forEach((d) => d())];
  } finally {
    scope = prev;
    active = prevActive;
  }
}

/** Registers a cleanup with the current scope. */
export function onCleanup(fn: () => void) {
  scope?.push(fn);
}
