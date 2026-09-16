package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urborg, Tomb of Yawgmoth — "Each land is a Swamp in addition to
// its other land types."
//
// That sentence is the whole card, and it is one Layer-4 static
// ability. It has nonetheless been uncatalogable for four sprints:
// #258 wrote Urborg up and then declined to ship it, because the
// static "would apply cleanly and do nothing." The layer engine
// computed the added Swamp type, the wire carried it, and then the
// only thing that turns a land type into mana — the synthetic
// basic-land ability behind ManaAbilitiesForCard — read the PRINTED
// type line and never saw it.
//
// Two changes in this branch make the sentence real:
//
//   - Card.IsLand and the rest of the type predicates read
//     Effective().Types rather than Card.TypeLine, so a Layer-4
//     type change reaches the engine instead of stopping at the
//     wire.
//   - The intrinsic land-type mana abilities are derived from
//     EFFECTIVE subtypes and keyed off the basic land TYPE rather
//     than the Basic supertype, which is what CR 305.6 has always
//     said. Urborg never grants the supertype — it grants the type
//     — so the old rule could not have worked for it even in
//     principle.
//
// Consequences worth knowing at the table, all of them the printed
// card's real behaviour:
//
//   - Every land on the battlefield, under every player's control,
//     taps for {B} on top of whatever it already made. That
//     includes Urborg itself, which is a land and therefore a
//     Swamp, so Urborg alone taps for {B}.
//   - "In addition to its other land types" means nothing is lost:
//     a Forest still makes {G}, and it keeps its printed ability
//     in ability slot 0, so the auto-tapper's first-ability choice
//     for an untouched land is exactly what it was.
//   - Cabal Coffers and Magus of the Coffers count Swamps, and
//     every land now is one. Neither is in the catalog yet; when
//     one lands, it reads Effective().Subtypes and gets this for
//     free.
//
// Known narrowing: a land whose mana ability is declared by a
// catalog Spec keeps that declaration, and ManaAbilitiesForCard
// only appends the intrinsic colours the declaration cannot
// already make. So Urborg does give a Sunpetal Grove its {B}, but
// a hypothetical Spec that deliberately declared a land's ability
// as something other than its land types would shadow the type
// half. No card in the catalog is in that position today.
//
// DECLARED GAP, CR 613.8: "each land" reads a type that other layer-4
// effects write, so Urborg depends on any effect that makes something
// a land and should apply after it whatever the timestamps. The layer
// engine orders layer 4 by timestamp only, so a permanent made a land
// by an effect NEWER than Urborg is not a Swamp. The catalog pairs
// today are Song of the Dryads attached after Urborg entered and
// Arixmethes, Slumbering Isle entering after Urborg; both are pinned,
// skipped, in layer_dependency_pairs_test.go. The caveat goes when
// dependency ordering lands.
func init() {
	Register(Spec{
		OracleID:     "db6174d7-211d-4817-b8e4-8384594c83f9",
		Name:         "Urborg, Tomb of Yawgmoth",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A permanent that becomes a land after Urborg is already on the battlefield — such as one enchanted by a later Song of the Dryads, or an Arixmethes, Slumbering Isle that entered after Urborg — isn't a Swamp and doesn't tap for {B}.",
		},
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				// "Each land" — every land on the battlefield, not
				// just the controller's. Reading IsLand through the
				// effective view is deliberate: a permanent that
				// some other Layer-4 effect has made a land is one.
				// Timestamp order should NOT decide which of the two
				// saw the other: under CR 613.8a Urborg depends on
				// any effect that makes something a land, so it
				// applies after that effect whichever entered first.
				// The engine orders layer 4 by timestamp only, so a
				// Song of the Dryads or an Arixmethes newer than
				// Urborg leaves its permanent without Swamp (#668).
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsLand()
				},
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					// Idempotent, and appended rather than
					// prepended so a printed land type keeps its
					// slot in the derived ability list.
					for _, st := range c.Subtypes {
						if st == "Swamp" {
							return
						}
					}
					c.Subtypes = append(c.Subtypes, "Swamp")
				},
			},
		},
	})
}
