// modelKey.ts — your own model key for AI opponents (ADR 0182): the
// Settings › AI opponents tab over GET / PUT / DELETE /me/model-key and
// POST /me/model-key/test, and the store the tab and the lobby's bot
// picker share.
//
// The key goes to the server once, on save, and never comes back: the
// server answers with the provider, the model, the last four characters
// and dates. Nothing here keeps the key after the request that carries
// it. A 503 (the server has no sealing key, so the feature is off) and a
// 403 (not a signed-in person) both mean "not available here", and the
// tab hides itself.

import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { signedInUserID } from "./myGames";
import { authFetch, LobbyApiError, session } from "./session";

/** One provider in the picker (modelkeys.Provider). */
export interface ModelProvider {
  id: string;
  label: string;
  transport: "anthropic" | "openai";
  base_url?: string;
  model_hint?: string;
}

/** A saved key as the server shows it (modelkeys.View): never the key. */
export interface ModelKeyView {
  provider: string;
  provider_label: string;
  base_url?: string;
  model: string;
  last4: string;
  games_per_day: number;
  calls_per_game: number;
  games_today: number;
  created_at: string;
  updated_at: string;
  last_used_at?: string;
  last_error?: string;
  last_error_at?: string;
}

export interface ModelKeyLimits {
  min_games_per_day: number;
  max_games_per_day: number;
  default_games_per_day: number;
  min_calls_per_game: number;
  max_calls_per_game: number;
  default_calls_per_game: number;
}

/** The body of GET, PUT and DELETE /me/model-key. */
export interface ModelKeyState {
  available: boolean;
  key: ModelKeyView | null;
  providers: ModelProvider[];
  limits: ModelKeyLimits;
}

export const DEFAULT_LIMITS: ModelKeyLimits = Object.freeze({
  min_games_per_day: 1,
  max_games_per_day: 100,
  default_games_per_day: 10,
  min_calls_per_game: 1,
  max_calls_per_game: 5000,
  default_calls_per_game: 500,
});

/** OFF is what every surface shows until the server says otherwise. */
export const MODEL_KEY_OFF: ModelKeyState = Object.freeze({
  available: false,
  key: null,
  providers: [],
  limits: DEFAULT_LIMITS,
});

export const CUSTOM_PROVIDER = "custom";

export const modelKeyState: Writable<ModelKeyState> = guardedWritable(
  MODEL_KEY_OFF,
  "modelKeyState",
);

/** normalizeModelKeyState turns whatever the server sent into a state. */
export function normalizeModelKeyState(raw: unknown): ModelKeyState {
  const r = (raw ?? {}) as Partial<ModelKeyState>;
  if (r.available !== true) return MODEL_KEY_OFF;
  return {
    available: true,
    key: r.key && typeof r.key === "object" ? (r.key as ModelKeyView) : null,
    providers: Array.isArray(r.providers) ? r.providers : [],
    limits: { ...DEFAULT_LIMITS, ...(r.limits ?? {}) },
  };
}

/** loadModelKey reads the state into the store; any refusal leaves it OFF. */
export async function loadModelKey(): Promise<ModelKeyState> {
  if (signedInUserID(get(session)) === null) {
    modelKeyState.set(MODEL_KEY_OFF);
    return MODEL_KEY_OFF;
  }
  try {
    const res = await authFetch("/me/model-key", { cache: "no-store" });
    const got = normalizeModelKeyState(await res.json());
    modelKeyState.set(got);
    return got;
  } catch (e) {
    if (!(e instanceof LobbyApiError && (e.status === 503 || e.status === 403))) {
      console.warn("could not read the model key settings", e);
    }
    modelKeyState.set(MODEL_KEY_OFF);
    return MODEL_KEY_OFF;
  }
}

/** The form's fields. An empty key keeps the saved one. */
export interface ModelKeyForm {
  provider: string;
  base_url: string;
  model: string;
  key: string;
  games_per_day: number;
  calls_per_game: number;
}

/** saveModelKey sends the form and resolves to the server's answer. */
export async function saveModelKey(form: ModelKeyForm): Promise<ModelKeyState> {
  const body: Record<string, unknown> = {
    provider: form.provider,
    model: form.model.trim(),
    games_per_day: form.games_per_day,
    calls_per_game: form.calls_per_game,
  };
  if (form.provider === CUSTOM_PROVIDER) body.base_url = form.base_url.trim();
  if (form.key.trim() !== "") body.key = form.key.trim();
  const res = await authFetch("/me/model-key", { method: "PUT", body: JSON.stringify(body) });
  const got = normalizeModelKeyState(await res.json());
  modelKeyState.set(got);
  return got;
}

/** removeModelKey deletes the saved key. */
export async function removeModelKey(): Promise<ModelKeyState> {
  const res = await authFetch("/me/model-key", { method: "DELETE" });
  const got = normalizeModelKeyState(await res.json());
  modelKeyState.set(got);
  return got;
}

export interface ModelKeyTestResult {
  ok: boolean;
  message: string;
}

/** testModelKey makes one small call with the saved key. */
export async function testModelKey(): Promise<ModelKeyTestResult> {
  const res = await authFetch("/me/model-key/test", { method: "POST" });
  const raw = (await res.json()) as { ok?: boolean; message?: string; key?: ModelKeyView };
  const prev = get(modelKeyState);
  if (raw.key && prev.available) modelKeyState.set({ ...prev, key: raw.key });
  return { ok: raw.ok === true, message: raw.message ?? "" };
}

/** formFor fills the form from a saved key, or the defaults. */
export function formFor(state: ModelKeyState): ModelKeyForm {
  const k = state.key;
  return {
    provider: k?.provider ?? state.providers[0]?.id ?? "anthropic",
    base_url: k?.base_url ?? "",
    model: k?.model ?? "",
    key: "",
    games_per_day: k?.games_per_day ?? state.limits.default_games_per_day,
    calls_per_game: k?.calls_per_game ?? state.limits.default_calls_per_game,
  };
}

/**
 * keyRequired reports whether saving the form needs a key typed in: no
 * key is saved, or the provider (or a custom address) changed, since a
 * key belongs to one provider.
 */
export function keyRequired(state: ModelKeyState, form: ModelKeyForm): boolean {
  const k = state.key;
  if (!k) return true;
  if (k.provider !== form.provider) return true;
  return form.provider === CUSTOM_PROVIDER && (k.base_url ?? "") !== form.base_url.trim();
}

/** describeKey is the one-line summary of a saved key. */
export function describeKey(k: ModelKeyView): string {
  return `${k.provider_label} · ${k.model} · key ending ${k.last4 || "…"}`;
}
