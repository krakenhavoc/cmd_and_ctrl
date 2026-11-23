package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b24_helpers.go — shared bodies for the reprint pass, batch 24.

// exileTargetArtifactOrEnchantmentSpell is the Spec of a spell whose whole
// text is "Exile target artifact or enchantment." (Fade into Antiquity),
// named once so it is not a second copy of Altar's Light's body (the clone
// gate, TestNoNewExactClonesInTheCatalog).
func ExileTargetArtifactOrEnchantmentSpell(oracleID, name string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	}
}

// freeCastFromDamagedPlayersGraveyard is "Whenever <this> deals combat damage
// to a player, you may cast target instant or sorcery card from that
// player's graveyard without paying its mana cost. If that spell would be
// put into a graveyard, exile it instead." (Deluxe Dragster, Wrexial). The
// target clause names the damaged player's graveyard (Fire Nation
// Salvagers' TargetsFrom); the cast is Arcane Heist's free, exile-instead
// grant offered during the resolution (CR 608.2g).
func FreeCastFromDamagedPlayersGraveyard(name string) game.TriggeredAbility {
	damage := WheneverThisDealsCombatDamageToAPlayer(
		name+" — you may cast target instant or sorcery card from that player's graveyard without paying its mana cost",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					Cmdr2S42FreeGraveyardCast(ctx, name, FreeInstantOrSorcery(), true, []uuid.UUID{t.ID})
					break
				}
			}
			return nil
		})
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card from that player's graveyard",
			Or(Instant(), Sorcery()), OwnedBy(tc.Event.Target))
	}
	return damage
}

// zurzothDevilAttackedPlayers is "those players" of Zurzoth's second
// ability: the players (not planeswalkers or battles, CR 508.3a) that the
// controller's Devils were declared as attackers against in the event
// batch ev belongs to, in the order the attacks were declared, each once.
// Read off the log, so a Devil that has died or left combat since still
// counts, and a Devil is judged as it last stood on the battlefield when
// it is gone (CR 603.10, 608.2h).
//
// Caller must hold g.mu.
func ZurzothDevilAttackedPlayers(g *game.Game, ev game.Event, controller uuid.UUID, devil string) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Events {
		e := g.Events[i]
		if e.Batch != ev.Batch || e.Kind != game.EventAttack || e.Actor != controller {
			continue
		}
		if !AttackerHadSubtype(g, e.CardID, devil) {
			continue
		}
		p := AttackedPlayer(g, e)
		if p == uuid.Nil || slices.Contains(out, p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// attackerHadSubtype reports whether the attacking object has the creature
// type `sub` (a word of the source's text, CR 612.2): read live while it is
// on the battlefield, else off its last-known record.
func AttackerHadSubtype(g *game.Game, id uuid.UUID, sub string) bool {
	if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneBattlefield {
		if c, ok := g.LookupCardForEffect(id); ok {
			return c.HasSubtype(sub)
		}
	}
	info, ok := g.LastKnownPermanentForEffect(id)
	return ok && (info.Characteristic.AllCreatureTypes || slices.Contains(info.Characteristic.Subtypes, sub))
}

// putAFlyingCounterOnTheTarget is "put a flying counter on target creature":
// the one legal target of the ability or trigger gets the counter (Hexavus,
// Wingspan Mentor).
func PutAFlyingCounterOnTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterFlying, N: 1}.Apply(ctx)
}
