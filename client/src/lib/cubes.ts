// cubes.ts — a signed-in person's saved cubes (ADR 0222 §8, #5441): the
// /me/cubes routes and the words the pages say about a cube's check.
//
// A cube is a named list of cards a draft can deal its packs from in
// place of a set's. Mirrors server/internal/cubes (Report) and
// server/internal/lobby/cubes_http.go; docs/lobby.md "Cubes" is the
// contract. Hand-maintained: update both sides in lockstep.

import { authFetch } from "./session";

/** What the server's check of a cube found. Mirrors cubes.Report. */
export interface CubeCheck {
  /** How many cards the list holds, copies counted. */
  cards: number;
  /** Names the card index does not know. */
  unknown: string[];
  /** Cards that are not fully built yet. */
  not_full: string[];
  /** Cards a singleton cube holds more than once. */
  repeated: string[];
  /** The list is over the most a cube may hold. */
  too_large: boolean;
  /** No card blocks it. Whether it is large enough depends on the draft. */
  ready: boolean;
  /** The largest draft of 15-card packs it can deal; 0 for none. */
  max_seats: number;
}

/** One cube. `list` is present on the routes that answer with one cube. */
export interface Cube {
  id: string;
  name: string;
  singleton: boolean;
  list?: string;
  check: CubeCheck;
  updated_at: string;
}

/** The limits `GET /me/cubes` states beside the list. */
export interface CubeLimits {
  max_cubes: number;
  max_cards: number;
  max_name_length: number;
  pack_size: number;
  min_pack_size: number;
  max_pack_size: number;
  packs_per_seat: number;
}

export interface MyCubes extends CubeLimits {
  cubes: Cube[];
}

/** The limits to state before the server has answered. Mirrors cubes.go and draft/source.go. */
export const CUBE_LIMITS: CubeLimits = {
  max_cubes: 20,
  max_cards: 1080,
  max_name_length: 60,
  pack_size: 15,
  min_pack_size: 9,
  max_pack_size: 20,
  packs_per_seat: 3,
};

/** POST /draft's refusal for a cube from a caller who is not signed in. */
export const CUBE_NEEDS_SIGN_IN_CODE = "cube_needs_sign_in";

/** The hash of the My cubes page. */
export const MY_CUBES_HASH = "#/cubes";

/** A cube's list comes from a pasted list or from a deck link, never both. */
export interface SaveCubeInput {
  name?: string;
  text?: string;
  url?: string;
  singleton?: boolean;
}

function saveBody(input: SaveCubeInput): string {
  const body: Record<string, unknown> = {};
  const name = input.name?.trim();
  if (name) body.name = name;
  const text = input.text?.trim();
  if (text) body.text = text;
  const url = input.url?.trim();
  if (url) body.url = url;
  if (input.singleton !== undefined) body.singleton = input.singleton;
  return JSON.stringify(body);
}

/** `GET /me/cubes`: the caller's cubes by name, each with its check, without lists. */
export async function fetchMyCubes(signal?: AbortSignal): Promise<MyCubes> {
  const res = await authFetch("/me/cubes", { signal });
  const body = (await res.json()) as Partial<MyCubes>;
  return { ...CUBE_LIMITS, ...body, cubes: body.cubes ?? [] };
}

const cubePath = (id: string): string => `/me/cubes/${encodeURIComponent(id)}`;

/** `GET /me/cubes/{id}`: one cube with its list, checked against the catalog as it is now. */
export async function fetchMyCube(id: string, signal?: AbortSignal): Promise<Cube> {
  const res = await authFetch(cubePath(id), { signal });
  return ((await res.json()) as { cube: Cube }).cube;
}

/** `POST /me/cubes`. */
export async function createCube(input: SaveCubeInput): Promise<Cube> {
  const res = await authFetch("/me/cubes", { method: "POST", body: saveBody(input) });
  return ((await res.json()) as { cube: Cube }).cube;
}

/** `PUT /me/cubes/{id}`: what the input leaves out stays as it was. */
export async function updateCube(id: string, input: SaveCubeInput): Promise<Cube> {
  const res = await authFetch(cubePath(id), { method: "PUT", body: saveBody(input) });
  return ((await res.json()) as { cube: Cube }).cube;
}

/** `DELETE /me/cubes/{id}`. */
export async function deleteCube(id: string): Promise<void> {
  await authFetch(cubePath(id), { method: "DELETE" });
}

/** How many cards a draft deals: three packs a seat. */
export function cubeNeed(
  seats: number,
  packSize: number,
  packsPerSeat = CUBE_LIMITS.packs_per_seat,
): number {
  return seats * packsPerSeat * packSize;
}

/** Whether a cube can deal a draft of this many seats and this pack size. */
export function cubeFits(c: Pick<Cube, "check">, seats: number, packSize: number): boolean {
  return c.check.ready && c.check.cards >= cubeNeed(seats, packSize);
}

/** "1 card", "360 cards". */
export function cubeCards(n: number): string {
  return n === 1 ? "1 card" : `${n} cards`;
}

/** How many cards block a cube, each name counted once. */
export function cubeBlockerCount(check: CubeCheck): number {
  return check.unknown.length + check.not_full.length + check.repeated.length;
}

/** A cube's state in a few words, for the list and the create form's picker. */
export function cubeStatus(check: CubeCheck): string {
  if (check.cards === 0) return "No cards";
  if (check.too_large) return "Too many cards";
  const blocked = cubeBlockerCount(check);
  if (blocked > 0) return blocked === 1 ? "1 card blocks it" : `${blocked} cards block it`;
  if (check.max_seats === 0) return "Too small to draft";
  return `Ready for up to ${check.max_seats} seats`;
}

/**
 * Why a ready cube cannot deal the draft the form describes, as a
 * sentence; "" when it can.
 */
export function cubeTooSmallText(
  c: Pick<Cube, "name" | "check">,
  seats: number,
  packSize: number,
): string {
  const need = cubeNeed(seats, packSize);
  if (c.check.cards >= need) return "";
  return `${c.name} has ${cubeCards(c.check.cards)}. ${seats} seats with packs of ${packSize} need ${need}: choose fewer seats or smaller packs.`;
}

/** The pack sizes the create form offers. */
export function packSizeChoices(
  limits: Pick<CubeLimits, "min_pack_size" | "max_pack_size">,
): number[] {
  const out: number[] = [];
  for (let n = limits.min_pack_size; n <= limits.max_pack_size; n++) out.push(n);
  return out;
}

// One row of a stored list: "2 Name" or "2 Name (SET) 123".
const ROW = /^\s*\d+\s+(.+?)(?:\s+\([A-Za-z0-9]+\)(?:\s+\S+)?)?\s*$/;

/**
 * A cube's list without the rows that name one of `names`: what "drop
 * the cards that block" saves. Names are matched whole, ignoring case.
 */
export function dropCards(list: string, names: readonly string[]): string {
  const gone = new Set(names.map((n) => n.trim().toLowerCase()));
  return list
    .split("\n")
    .filter((line) => {
      if (!line.trim()) return false;
      const name = ROW.exec(line)?.[1] ?? line.trim();
      return !gone.has(name.toLowerCase());
    })
    .join("\n");
}
