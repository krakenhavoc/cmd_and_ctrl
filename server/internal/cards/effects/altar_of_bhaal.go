package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altar of Bhaal // Bone Offering — an adventure card (CR 715), #1600:
//
//	Altar of Bhaal — Artifact {1}{B}
//	  "{2}{B}, {T}, Exile a creature you control: Return target
//	   creature card from your graveyard to the battlefield. Activate
//	   only as a sorcery."
//	Bone Offering — Sorcery — Adventure {2}{B}
//	  "Create a tapped 4/1 black Skeleton creature token with menace.
//	   (Then exile this card. You may cast the artifact later from
//	   exile.)"
//
// Cauldron of Essence's reanimation (cauldron_of_essence.go) with the
// cost component the seam row named this card for: "Exile a creature
// you control" (game.ExilePermanentsCost, ADR 0020's 2026-10-03
// amendment) instead of a sacrifice. The creature is exiled at
// announce, so it is never in the graveyard to be its own target, and
// nothing dies — a Blood Artist does not drain off it. The target is
// validated before the cost is paid (CR 601.2c before 601.2h), exactly
// as Cauldron's is, and the card returns under its owner's control,
// which "from your graveyard" makes the activator.
//
// Two keys, one card, as Foulmire Knight's: the artifact keeps the
// bare oracle ID and the Adventure takes "<oracle_id>#1". The
// adventure lifecycle (cast it, exile it, cast the artifact later from
// exile) is the engine's CR 715 branch and nothing here mentions it.
//
// No simplification.
const altarOfBhaalOracleID = "0a364b66-95df-480b-a733-e90f6d5c4d2b"

func init() {
	Register(Spec{
		OracleID:     altarOfBhaalOracleID,
		Name:         "Altar of Bhaal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}, {T}, Exile a creature you control: Return target creature card from your graveyard to the battlefield.",
			Cost:         Plus(ManaCost("{2}{B}"), TapCost(), ExileACreatureYouControl()),
			Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			SorcerySpeed: true,
			Effect:       returnFirstLegalGraveyardTargetToBattlefield,
		}},
	})

	Register(Spec{
		OracleID:     altarOfBhaalOracleID + "#1",
		Name:         "Bone Offering",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateTokenAdvanced{
				Spec: Token(TokenCard("4/1 black Skeleton with menace")).EntersTapped(),
				N:    1,
			}.Apply(ctx)
		},
	})
}
