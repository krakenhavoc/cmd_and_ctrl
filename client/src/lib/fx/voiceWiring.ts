// Glue between the game log and the voice bus (ADR 0209 §5, #4398): a cue
// names a card, the client catalog says which oracle_id that is, the bus
// speaks. With an empty manifest nothing here touches the network.

import { fetchCatalog, type CatalogResponse } from "../catalog";
import type { VoiceCue } from "./voiceCues";
import { voices, type VoiceBus } from "./voices";

let byName: Promise<Map<string, string>> | null = null;

function index(res: CatalogResponse): Map<string, string> {
  const m = new Map<string, string>();
  for (const e of res.cards ?? []) {
    m.set(e.name.toLowerCase(), e.oracle_id);
    for (const f of e.faces ?? []) m.set(f.name.toLowerCase(), e.oracle_id);
  }
  return m;
}

/** oracleIDForName maps a card name to its oracle_id through the client catalog. */
export function oracleIDForName(
  name: string,
  load: () => Promise<CatalogResponse> = fetchCatalog,
): Promise<string | undefined> {
  byName ??= load().then(index, (err: unknown) => {
    byName = null; // try again at the next cue
    throw err;
  });
  return byName.then((m) => m.get(name.toLowerCase())).catch(() => undefined);
}

/** forgetCatalog drops the cached name index (tests). */
export function forgetCatalog(): void {
  byName = null;
}

/** speakCue plays the line a cue names, if its walker has one. */
export async function speakCue(
  cue: VoiceCue,
  bus: VoiceBus = voices,
  load?: () => Promise<CatalogResponse>,
): Promise<void> {
  if (bus.empty) return;
  const id = await oracleIDForName(cue.name, load);
  if (id) bus.speak(id, cue.moment);
}
