package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_swift_savior.go — Aang, Swift Savior // Aang and La, Ocean's
// Fury. A `transform` double-faced legendary creature, and the
// commander behind issues #325 and #343.
//
//	Aang, Swift Savior            {1}{W}{U}   2/3
//	Legendary Creature — Human Avatar Ally
//	  Flash
//	  Flying
//	  When Aang enters, airbend up to one other target creature or
//	  spell. (Exile it. While it's exiled, its owner may cast it for
//	  {2} rather than its mana cost.)
//	  Waterbend {8}: Transform Aang.
//
//	Aang and La, Ocean's Fury                  5/5
//	Legendary Creature — Avatar Spirit Ally
//	  Reach, trample
//	  Whenever Aang and La attack, put a +1/+1 counter on each
//	  tapped creature you control.
//
// # Why this could not be written before ADR 0034
//
// Nothing stopped Register from taking this oracle ID. What stopped
// the card from working is that it never reached the engine in a
// playable state. Scryfall puts NOTHING at the top level for a
// transform printing — mana_cost, colors, power and toughness are
// all null, and the type line is the concatenation — and
// deck.toGameCard copied exactly those. So Aang imported as a
// costless, colourless 0/0, which is #343's "he shows as a 0/0", and
// the commander-identity half of #276. A spec attached to that would
// have been an ETB trigger on a card nobody could sensibly cast.
//
// The face model fixes the import, so the card is now a {1}{W}{U}
// 2/3 with flash and flying, and this is the trigger #325 reported
// missing.
//
// # Waterbend {8}
//
// "Waterbend {8}: Transform Aang" is an activated ability whose cost
// is a waterbend cost (CR 701.67a): pay {8}, and for each of those
// eight generic mana you may tap an untapped artifact or creature you
// control instead. ADR 0079 built the TRANSFORM half —
// `Game.TransformPermanentForEffect` / `effects.TransformThis`, with
// the CR 712.18 hygiene (counters, damage, attachments and the CR
// 613.7 timestamp all surviving the flip) — and #1310 built the COST:
// `WaterbendCost("{8}")` puts the {8} in the mana component and the
// tap clause beside it (`game.AbilityCost.Waterbend`). Until then the
// ability charged a flat {8} in mana, which was weaker than printed.
//
// Two things the rules give the ability for free, and the engine
// honours both: tapping to waterbend is not the {T} symbol, so a
// creature that arrived this turn may help (CR 302.6); and the cost
// prints no {T}, so Aang himself — flash, so often freshly arrived —
// may be one of the eight.
//
// The back face's attack trigger (aang_and_la_oceans_fury.go) ships
// in full and is reached through this ability.
//
// # The ETB clause
//
// "Up to one OTHER target creature or spell" is two zones in one
// clause, which TargetSpec expresses directly: battlefield for the
// creature, stack for the spell. Min 0 makes "up to one" a real
// choice — announcing no target is legal, and CR 601.2c is satisfied
// either way.
//
// Airbending a SPELL wedged the table until #1318: the exile route
// moved the card and left its stack record behind, so nothing could
// resolve past it. The route now retires the record for every card
// that leaves the stack (ADR 0013 §5ad), and
// TestAangSwiftSaviorAirbendsASpell plays it through.
//
// "Other" excludes Aang himself, and that matters more here than it
// does on Aang, the Last Airbender: this Aang has FLASH, so he can
// enter in response to a spell and would otherwise be a legal target
// for his own trigger.
func init() {
	Register(Spec{
		// Face 0's catalog key is the BARE oracle ID — see
		// game.CatalogKey. The back face registers under
		// aangSwiftSaviorOracle + "#1" — see
		// aang_and_la_oceans_fury.go.
		OracleID:        aangSwiftSaviorOracle,
		Name:            "Aang, Swift Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Activated: []ActivatedAbility{{
			Label:  "Waterbend {8}: Transform Aang.",
			Cost:   WaterbendCost("{8}"),
			Effect: Do(TransformThis{}),
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: targetOtherCreatureOrSpell(
				"up to one other target creature or spell",
			),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Aang, Swift Savior — airbend a creature or spell",
					AirbendOtherTarget)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Aang, Swift Savior — airbend a creature or spell? " +
					"(Exile it; its owner may cast it for {2}.)",
			},
		}},
	})
}

// targetOtherCreatureOrSpell is "up to one target creature or
// spell": a creature on the battlefield, or any spell on the stack.
// The "other" is enforced at resolution — see the note there.
//
// Mode is "any" rather than "creature" or "stack_spell" because the
// client's Mode string decides which SURFACES enter targeting — a
// clause spanning the battlefield and the stack needs both, and
// legality still comes from CardOK rather than from Mode. Players is
// left false, so "any" does not let this point at a player.
func targetOtherCreatureOrSpell(label string) *game.TargetSpec {
	isCreature := Creature()
	return &game.TargetSpec{
		Mode:  "any",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneBattlefield, game.ZoneStack},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, zone game.ZoneKind) bool {
			if zone == game.ZoneStack {
				// Every card on the stack is a spell (abilities live
				// in StackMeta, not as cards), so no further
				// narrowing is needed or correct: "spell" here means
				// any spell, not any creature spell.
				return true
			}
			return isCreature(g, caster, c)
		},
		// "Up to one" — Min 0 is what makes declining a legal
		// announcement rather than a failed cast.
		Min: 0, Max: 1,
	}
}

// aangSwiftSaviorOracle is shared with the back face
// (aang_and_la_oceans_fury.go), which registers under it plus "#1"
// (game.CatalogKey).
const aangSwiftSaviorOracle = "cbf09050-39d0-463b-96db-9e22011ae0d8"
