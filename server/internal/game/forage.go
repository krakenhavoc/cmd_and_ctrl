package game

import "github.com/google/uuid"

// forage.go — #5184: the forage keyword action (CR 701.61a). "To forage
// means 'Exile three cards from your graveyard or sacrifice a Food.'"
// Bloomburrow prints it three ways:
//
//	a cost         "{2}, Forage: Put a +1/+1 counter on each other
//	               Squirrel you control."      Camellia, the Seedmiser
//	an additional  "As an additional cost to cast this spell, forage
//	cost branch    or pay {B}."                 Feed the Cycle
//	an effect      "When this creature enters, you may forage. If you
//	               do, draw a card."           Treetop Sentries
//
// and reads it once more: "Whenever you forage, put a +1/+1 counter on
// this creature" (Corpseberry Cultivator).
//
// Forage is not a new payment. It is an either/or between two payments
// each owner already has — exiling cards from your own graveyard
// (ExileCost, PayActionExileFromGraveyard) and sacrificing a permanent
// (the sacrifice clause, PayActionSacrifice) — so each owner builds it
// from its own either/or machinery and marks it as a forage:
//
//	activated ability   AbilityCost.Forage over ForageAbilityBranches,
//	                    folded by ChosenAbilityCost like any OrCost
//	additional cost     two AdditionalCost.Either branches with Forage
//	                    set (effects.ForageOrPay)
//	effect              a PayActionEither with Forage set
//	                    (ForagePayAction), asked as a may-pay
//
// That keeps ONE gate per owner, the one each of its either/or costs
// already has: a branch is payable when its walk finds enough (three
// cards in the graveyard, or a Food), so forage is payable exactly when
// one of the two is possible (CR 118.3), and the activation, the view's
// branch list and the enumerator read the same walk. A forage that
// neither branch can pay can't be chosen: the ability can't be
// activated, the spell's forage branches are greyed, and a "you may
// forage" asks nothing.
//
// The mark is what the rest of the game reads. Paying a marked cost
// records PaidCost.Foraged ("if you foraged") and emits EventForage
// (noteForageLocked), which "whenever you forage" watches. The cards
// exiled and the Food sacrificed are recorded where those payments
// always record them (PaidCost.Exiled, PaidCost.SacrificedObjects).

// ForageExileCount is how many cards from your graveyard a forage
// exiles (CR 701.61a).
const ForageExileCount = 3

// FoodQuery is the permanents a forage may sacrifice: a Food, read off
// its effective characteristics, so a Food token, a Food artifact and a
// Food artifact creature all count (CR 205.3g: Food is an artifact
// type).
func FoodQuery() PermanentQuery {
	return PermanentQuery{Subtypes: []string{"Food"}}
}

// ForageAbilityBranches is forage as the two branches of an activated
// ability's either/or cost (AbilityCost.Either): exile three cards from
// your graveyard, or sacrifice a Food. The order is the printed one, and
// the index is what `cost_branch` names.
func ForageAbilityBranches() []AbilityCostBranch {
	return []AbilityCostBranch{
		{Kind: AbilityCostBranchExileGraveyard, Count: ForageExileCount, Label: "three cards"},
		{Kind: AbilityCostBranchSacrifice, Count: 1, Label: "a Food", Of: FoodQuery()},
	}
}

// ForagePayAction is forage as a payment made while something resolves
// ("you may forage. If you do, …"): one either payment whose answer
// names the branch (ADR 0187 §5), marked Forage so paying either branch
// is a forage. The words are the printed reminder's.
func ForagePayAction() *PayAction {
	return &PayAction{
		Kind:   PayActionEither,
		Forage: true,
		Either: []PayAction{
			{Kind: PayActionExileFromGraveyard, Count: ForageExileCount, Words: "exile three cards from your graveyard"},
			{Kind: PayActionSacrifice, Count: 1, Of: FoodQuery(), Words: "sacrifice a Food"},
		},
	}
}

// noteForageLocked records that `player` foraged, paying for `source`:
// one EventForage, which "whenever you forage" (Corpseberry Cultivator)
// watches. Called by each owner once its forage has been paid. Caller
// must hold g.mu in write mode.
func (g *Game) noteForageLocked(player, source uuid.UUID) {
	g.EmitEvent(Event{Kind: EventForage, Actor: player, Source: source})
}
