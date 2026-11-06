package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azula, Cunning Usurper — Legendary Creature — Human Noble Rogue
// {2}{U}{B}{B}, 4/4 (Avatar: The Last Airbender, #4459):
//
//	"Firebending 2 (Whenever this creature attacks, add {R}{R}. This mana
//	 lasts until end of combat.)
//	 When Azula enters, target opponent exiles a nontoken creature they
//	 control, then they exile a nonland card from their graveyard.
//	 During your turn, you may cast cards exiled with Azula and you may
//	 cast them as though they had flash. Mana of any type can be spent to
//	 cast those spells."
//
// Firebending is Avatar Roku's (`firebending`, #3172): an attack trigger
// that uses the stack, adding red that lasts until end of combat.
//
// The enters trigger is two exile edicts in a row, each chosen by the
// targeted opponent (Doomfall's msc17ExileEdict shape): first one of the
// nontoken creatures they control, then, once that exile has landed, one
// of the nonland cards in their graveyard. Both exiles are LINKED to this
// Azula object (CR 607.2a, Card.ExiledWith), so "cards exiled with Azula"
// is exactly those two and no card another effect exiled. An opponent with
// no such creature, or no such card, is not asked that half.
//
// The cast clause is Null Summoner's standing permission over exile
// (ExiledWithSource), gated on "during your turn" (DuringYourTurn), with
// TimingFlash for "as though they had flash" and AnyType for "mana of any
// type". It is derived off the battlefield on every query, so it ends the
// moment this Azula leaves. The cards' own costs are paid, alternative
// costs included (the ruling). Casting a card moves it out of exile, so
// each is cast once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c9e351d-2e6c-4050-9c3d-b15bb9255a4d",
		Name:         "Azula, Cunning Usurper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Azula, Cunning Usurper — firebending 2", Do(firebending(2))),
			Targeting(
				WhenThisEnters("Azula, Cunning Usurper — target opponent exiles a nontoken creature they control, then a nonland card from their graveyard",
					azulaCunningUsurperEnters),
				TargetPlayer("target opponent", Opponent()),
			),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: DuringYourTurn(),
			Permission: game.CastPermission{
				Zone:     game.ZoneExile,
				Filter:   game.PermissionFilter{ExiledWithSource: true},
				AnyType:  true,
				CastOnly: true,
				Timing:   game.TimingFlash,
				Label:    "During your turn, cast a card exiled with Azula as though it had flash; mana of any type can be spent (Azula, Cunning Usurper)",
			},
		}},
	})
}

// azulaCunningUsurperEnters asks the targeted opponent for the creature.
func azulaCunningUsurperEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	var cards []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == opp && c.IsCreature() && !c.IsToken() {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return azulaCunningUsurperGraveyardPick(g, opp, ref)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: opp,
		Source:     ref.ID,
		Question:   "Azula, Cunning Usurper — exile a nontoken creature you control",
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    azulaCunningUsurperCreatureChosen,
		LiveZone:   game.ZoneBattlefield,
		Carry:      []uuid.UUID{ref.ID},
		CarryNums:  []int{ref.Epoch},
	})
	return nil
}

// azulaCunningUsurperRef reads the Azula object a prompt carries.
func azulaCunningUsurperRef(carry []uuid.UUID, nums []int) (game.ObjectRef, bool) {
	if len(carry) == 0 || len(nums) == 0 {
		return game.ObjectRef{}, false
	}
	return game.ObjectRef{ID: carry[0], Epoch: nums[0]}, true
}

// azulaCunningUsurperCreatureChosen exiles the chosen creature linked to
// Azula, then asks for the graveyard card. On-disk key.
var azulaCunningUsurperCreatureChosen = game.RegisterChooseCardsThen("choose-cards/azula-cunning-usurper-creature",
	func(g *game.Game, r game.CardsChosen) error {
		ref, ok := azulaCunningUsurperRef(r.Carry, r.Nums)
		if !ok {
			return nil
		}
		then := game.KeyedTail{Ref: azulaCunningUsurperThen, Player: r.Chooser, Source: ref.ID,
			Carry: []uuid.UUID{ref.ID}, Nums: []int{ref.Epoch}}
		if len(r.Picked) == 0 {
			return then.Run(g, nil)
		}
		return g.ExileLinkedKeyedForEffect(ref, r.Picked, then)
	})

// azulaCunningUsurperThen is "then they exile a nonland card from their
// graveyard", once the creature's exile has landed. On-disk key.
var azulaCunningUsurperThen = game.RegisterTail("azula-cunning-usurper/graveyard", func(g *game.Game, o game.TailOutcome) error {
	ref, ok := azulaCunningUsurperRef(o.Tail.Carry, o.Tail.Nums)
	if !ok {
		return nil
	}
	return azulaCunningUsurperGraveyardPick(g, o.Tail.Player, ref)
})

// azulaCunningUsurperGraveyardPick asks `opp` for a nonland card in their
// graveyard.
func azulaCunningUsurperGraveyardPick(g *game.Game, opp uuid.UUID, ref game.ObjectRef) error {
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var cards []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if !c.IsLand() {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: opp,
		Source:     ref.ID,
		Question:   "Azula, Cunning Usurper — exile a nonland card from your graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneGraveyard,
		ThenKey:    azulaCunningUsurperCardChosen,
		LiveZone:   game.ZoneGraveyard,
		Carry:      []uuid.UUID{ref.ID},
		CarryNums:  []int{ref.Epoch},
	})
	return nil
}

// azulaCunningUsurperCardChosen exiles the chosen graveyard card linked to
// Azula. On-disk key.
var azulaCunningUsurperCardChosen = game.RegisterChooseCardsThen("choose-cards/azula-cunning-usurper-graveyard",
	func(g *game.Game, r game.CardsChosen) error {
		ref, ok := azulaCunningUsurperRef(r.Carry, r.Nums)
		if !ok || len(r.Picked) == 0 {
			return nil
		}
		return g.ExileLinkedKeyedForEffect(ref, r.Picked, game.KeyedTail{})
	})
