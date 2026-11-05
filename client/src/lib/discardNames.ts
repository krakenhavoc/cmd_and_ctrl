// discardNames.ts — #4896: the "with different names" rule of a discard
// cost (Ormos, Archive Keeper's "Discard three cards with different
// names"), for the picker. The server marks the clause with
// `discard_cost_different_names` and refuses a set whose names repeat;
// these keep the picker from building one.
//
// Names are compared as the server compares them: case-insensitively,
// trimmed, and with a split card's two halves ("Fire // Ice") both
// counting (CR 709.4, CR 201.2a), so Fire // Ice and Fire share a name. A
// card with no name is never one of the set (CR 201.2b); the server does
// not offer one, and these treat it as clashing with everything.

interface Named {
  instance_id: string;
  name?: string;
}

// The names a card has, as comparison keys.
export function cardNameKeys(c: Named | undefined): string[] {
  return (c?.name ?? "")
    .split(" // ")
    .map((n) => n.trim().toLowerCase())
    .filter((n) => n !== "");
}

function clashes(a: string[], b: string[]): boolean {
  return a.length === 0 || b.length === 0 || a.some((n) => b.includes(n));
}

// Whether picking `id` beside the already `chosen` cards would repeat a
// name. A card already chosen does not clash with itself, so it can
// always be un-picked.
export function discardNameRepeats(options: Named[], chosen: readonly string[], id: string): boolean {
  if (chosen.includes(id)) return false;
  const mine = cardNameKeys(options.find((c) => c.instance_id === id));
  return options.some(
    (c) => c.instance_id !== id && chosen.includes(c.instance_id) && clashes(mine, cardNameKeys(c)),
  );
}

// Whether `need` of the options can be picked with no name repeated:
// the shortfall check. A small search, because a split card's two names
// can block a set that greedy picking would miss.
export function canPickDifferentNames(options: Named[], need: number): boolean {
  const keys = options.map(cardNameKeys).filter((k) => k.length > 0);
  if (need <= 0) return true;
  const held: string[][] = [];
  const walk = (from: number): boolean => {
    if (held.length === need) return true;
    for (let i = from; i < keys.length && keys.length - i >= need - held.length; i++) {
      if (held.some((h) => clashes(h, keys[i]))) continue;
      held.push(keys[i]);
      if (walk(i + 1)) return true;
      held.pop();
    }
    return false;
  };
  return walk(0);
}

// Whether the picked cards' names all differ.
export function discardNamesDiffer(options: Named[], chosen: readonly string[]): boolean {
  const picked = options.filter((c) => chosen.includes(c.instance_id)).map(cardNameKeys);
  for (let i = 0; i < picked.length; i++) {
    for (let j = i + 1; j < picked.length; j++) {
      if (clashes(picked[i], picked[j])) return false;
    }
    if (picked[i].length === 0) return false;
  }
  return true;
}
