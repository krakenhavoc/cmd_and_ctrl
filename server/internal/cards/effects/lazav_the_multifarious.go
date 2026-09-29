package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lazav, the Multifarious — Legendary Creature — Shapeshifter
// {U}{B}, 1/3:
//
//	"When Lazav enters, surveil 1. (Look at the top card of your
//	 library. You may put it into your graveyard.)
//	 {X}: Lazav becomes a copy of target creature card in your
//	 graveyard with mana value X, except its name is Lazav, the
//	 Multifarious, it's legendary in addition to its other types,
//	 and it has this ability."
//
// #1723's second seam: no target clause could be bound to an
// activated ability's own announced X before this (only a spell's
// could, #1559) — "with mana value X" needed both that binding AND an
// EXACT comparison, ManaValueAtMostX being "X or less". The clause is
// TargetCardInGraveyard(...).WithManaValueEqualsX(): the engine reads
// the X this ability announces (CR 602.2b) and re-reads it at
// resolution (CR 608.2b), so a target whose mana value changes between
// announce and resolution — Cloudstone Curio bouncing it back and
// forth, an artifact losing a cost-reducing static — is refused rather
// than copied as printed. X=0 is a legal announcement: a mana-value-0
// creature card is as valid a target as any other.
//
// The copy is a duration copy with NO stated duration (#1593,
// become_copy.go), exactly Lazav, Dimir Mastermind's shape one card
// over: it lasts until Lazav leaves the battlefield or copies again,
// and the engine keeps only the newest record. Unlike its Dimir
// Mastermind cousin this Lazav copies a card that is STILL in the
// graveyard when the ability resolves (it targets there, rather than
// reading "that card" off a trigger), so game.OwnPrintedValues is not
// needed — the resolution re-check already guarantees the card is a
// legal graveyard target, mana value included, at the moment it is
// read.
//
// The except clause is three edits (CR 707.9): the name, the
// Legendary supertype, and a grant of this ability as a named bundle,
// so a Lazav that has become a Hill Giant can still activate {X}
// again.
//
// No simplification.
const lazavTheMultifariousGrant = "lazav-the-multifarious/this-ability"

var lazavTheMultifariousAbility = ActivatedAbility{
	Label: "{X}: Lazav becomes a copy of target creature card in your graveyard with mana value X, except its name is Lazav, the Multifarious, it's legendary in addition to its other types, and it has this ability.",
	Cost:  ManaCost("{X}"),
	Targets: TargetCardInGraveyard("target creature card in your graveyard with mana value X",
		Creature(), YouOwn()).WithManaValueEqualsX(),
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		self := ctx.Source()
		return BecomeCopy{
			Targets:  []uuid.UUID{self},
			Of:       targets[0].ID,
			Duration: CopyIndefinite,
			Except: func(v *game.PrintedValues) {
				v.SetName("Lazav, the Multifarious")
				v.AddSupertype("Legendary")
				v.GrantAbility(lazavTheMultifariousGrant)
			},
			Label: "Lazav, the Multifarious — becomes a copy",
		}.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:     "c14bcef2-6d49-4430-86d0-e5a87ba442d9",
		Name:         "Lazav, the Multifarious",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				"Lazav, the Multifarious — surveil 1",
				func(g *game.Game, item *game.StackItem) error {
					return Surveil{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Grants: []AbilityGrant{{
			Key:       lazavTheMultifariousGrant,
			Text:      "{X}: Lazav becomes a copy of target creature card in your graveyard with mana value X, except its name is Lazav, the Multifarious, it's legendary in addition to its other types, and it has this ability.",
			Activated: []ActivatedAbility{lazavTheMultifariousAbility},
		}},
		Activated: []ActivatedAbility{lazavTheMultifariousAbility},
	})
}
