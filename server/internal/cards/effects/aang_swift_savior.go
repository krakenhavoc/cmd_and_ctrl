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
// # Declared simplifications
//
//   - "Waterbend {8}: Transform Aang" is NOT implemented. Transform
//     is a face change on a permanent already on the battlefield —
//     SetFace(1) plus a layer-staleness mark — and it needs a
//     `transform` action verb the protocol does not have yet. ADR
//     0034 sequences it as its own PR, deliberately, because CR 712
//     hygiene (counters, damage, auras and the CR 613 timestamp all
//     surviving the flip) is the interesting part and not this one.
//     The back face's data is imported and on the wire; only the
//     verb is missing. This is the remaining half of #343.
//   - The back face's attack trigger is likewise unwired, for the
//     same reason: nothing can reach face 1 yet.
//
// # The ETB clause
//
// "Up to one OTHER target creature or spell" is two zones in one
// clause, which TargetSpec expresses directly: battlefield for the
// creature, stack for the spell. Min 0 makes "up to one" a real
// choice — announcing no target is legal, and CR 601.2c is satisfied
// either way.
//
// "Other" excludes Aang himself, and that matters more here than it
// does on Aang, the Last Airbender: this Aang has FLASH, so he can
// enter in response to a spell and would otherwise be a legal target
// for his own trigger.
func init() {
	Register(Spec{
		// Face 0's catalog key is the BARE oracle ID — see
		// game.CatalogKey. The back face would register under
		// "cbf09050-…#1" when it has rules to run; today it has none
		// the engine can reach.
		OracleID:        "cbf09050-39d0-463b-96db-9e22011ae0d8",
		Name:            "Aang, Swift Savior",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Aang can't transform — the \"Waterbend {8}\" ability and the back face's attack trigger don't work."},
		PrintedKeywords: []string{"flash", "flying"},
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
					func(g *game.Game, item *game.StackItem) error {
						// "Up to one": no target is a legal
						// announcement and resolves as nothing.
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						target := item.Targets[0]
						// "Other". Enforced here rather than in
						// CardOK because the target predicate is not
						// given the ability's source — the same
						// convention Aang, the Last Airbender's
						// "another target nonland permanent" follows.
						// It matters more on this Aang: he has
						// FLASH, so he can enter in response to a
						// spell and would otherwise be a perfectly
						// legal target for his own trigger.
						if target.ID == item.SourceCardID {
							return nil
						}
						return Airbend{Target: target.ID}.Apply(NewContext(g, item))
					})
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
