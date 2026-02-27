import type { AppAction } from "./actions";
import type { AppState } from "@/domain/models";

export type StoreListener = (state: AppState, action: AppAction) => void;

export class AppStore {
  private state: AppState;
  private listeners: StoreListener[] = [];

  constructor(initialState: AppState, private readonly reducer: (state: AppState, action: AppAction) => AppState) {
    this.state = initialState;
  }

  getState(): AppState {
    return this.state;
  }

  dispatch(action: AppAction): void {
    this.state = this.reducer(this.state, action);
    this.listeners.forEach((listener) => listener(this.state, action));
  }

  subscribe(listener: StoreListener): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter((item) => item !== listener);
    };
  }
}
