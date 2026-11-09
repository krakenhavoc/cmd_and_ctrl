package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const etrataGrant = "etrata-deadly-fugitive/turn-face-up-or-cast"

// Etrata, Deadly Fugitive — Legendary Creature — Vampire Assassin {1}{U}{B}
// (Murders at Karlov Manor, #4469):
//
//	"Deathtouch
//	 Face-down creatures you control have '{2}{U}{B}: Turn this creature
//	 face up. If you can't, exile it, then you may cast the exiled card
//	 without paying its mana cost.'
//	 Whenever an Assassin you control deals combat damage to an opponent,
//	 cloak the top card of that player's library."
//
// The grant is a layer-6 static over face-down creatures you control (a
// face-down permanent has no abilities of its own, CR 708.2a, so the ability
// is the only one it has). Its effect turns the creature face up as an
// effect, paying nothing (TurnFaceUpForEffect, CR 708.8), and when that is
// refused (a face-down instant or sorcery, CR 701.40g, or something that
// forbids turning face up) exiles it and offers its card for a free cast
// during the resolution (CR 608.2g, X locked at 0). Cloak is the cloak seam
// (CloakTopOfLibrary, game/cloak.go): the top card of the damaged
// opponent's library, entering under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87957d09-7e22-40fe-8770-ff66d1934a38",
		Name:            "Etrata, Deadly Fugitive",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Grants: []AbilityGrant{{
			Key: etrataGrant,
			Activated: []ActivatedAbility{{
				Label:  "{2}{U}{B}: Turn this creature face up. If you can't, exile it, then you may cast the exiled card without paying its mana cost.",
				Cost:   ManaCost("{2}{U}{B}"),
				Effect: etrataTurnFaceUpOrCast,
			}},
			Text: "{2}{U}{B}: Turn this creature face up. If you can't, exile it, then you may cast the exiled card without paying its mana cost.",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.FaceDownIsPermanent() && target.IsCreature() && target.Controller == source.Controller
			}, etrataGrant), "Face-down creatures you control have \"{2}{U}{B}: Turn this creature face up. If you can't, exile it, then you may cast the exiled card without paying its mana cost.\""),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, etrataAssassinHitAnOpponent,
				"Etrata — cloak the top card of that player's library",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return CloakTopOfLibrary{LibraryOwner: item.Trigger.Event.Target}.Apply(NewContext(g, item))
				}),
		},
	})
}

// etrataAssassinHitAnOpponent is "whenever an Assassin you control deals
// combat damage to an opponent": the damage is combat damage to a player
// other than you, from a creature you control that is an Assassin. "Assassin"
// is a word of the source's text (CR 612.2).
func etrataAssassinHitAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) || ev.Target == source.Controller {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.HasSubtype(source.TextSubtype("Assassin"))
}

// etrataTurnFaceUpOrCast is the granted ability's effect, run with the
// face-down creature as the source (the object that has the ability).
func etrataTurnFaceUpOrCast(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if !onBattlefield(g, self) {
		return nil
	}
	if g.TurnFaceUpForEffect(self, item.Controller, self) {
		return nil
	}
	// "If you can't": exile it, then you may cast the exiled card without
	// paying its mana cost.
	perm := game.CastPermission{
		Player:      item.Controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      self,
		SourceName:  "Etrata, Deadly Fugitive",
		Label:       "Etrata, Deadly Fugitive — cast the exiled card without paying its mana cost",
	}
	return exileWithPermissionThenOffer(g, self, self, perm, "Etrata, Deadly Fugitive", game.MayCastKeywordFreeCast)
}
