package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grub, Storied Matriarch // Grub, Notorious Auntie — {2}{B} Legendary
// Creature — Goblin Warlock 2/1 // Legendary Creature — Goblin Warrior 2/1
// (Lorwyn Eclipsed, #4458):
//
//	Front: "Menace
//	        Whenever this creature enters or transforms into Grub, Storied
//	        Matriarch, return up to one target Goblin card from your
//	        graveyard to your hand.
//	        At the beginning of your first main phase, you may pay {R}. If
//	        you do, transform Grub."
//	Back:  "Menace
//	        Whenever Grub attacks, you may blight 1. If you do, create a
//	        tapped and attacking token that's a copy of the blighted
//	        creature, except it has "At the beginning of the end step,
//	        sacrifice this token."
//	        At the beginning of your first main phase, you may pay {B}. If
//	        you do, transform Grub."
//
// Brigid, Clachan's Heart's shape for the faces and the pay-to-transform
// triggers (PayToTransformAtYourFirstMain).
//
// The attack trigger's blight is the effect-side Blight (#4683, CR
// 701.68a): any creature you control, Grub herself included, chosen as the
// trigger resolves. "If you do" is a creature chosen; the copy is of the
// blighted creature (CR 701.68c), read once its counter has landed, which
// does not change its copiable values (CR 707.2). A creature the -1/-1
// counter leaves at 0 toughness is still there to be copied: state-based
// actions wait until the ability has finished resolving (the ruling). The
// sacrifice trigger is gained as part of the copying process (CR 707.9a,
// Electroduplicate's SacrificeThisTokenAtEndStepGrant), so it is the
// token's own and fires at the beginning of every end step it sees. The
// token was never declared as an attacker (CR 508.4): its controller
// chooses what it attacks, offered where Grub went first.
//
// No simplification.
func init() {
	const oracle = "180286e4-67fa-44d5-9e90-4a4e99c016a7"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Grub, Storied Matriarch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventTransform}, ThisEntersOrTransformsInto(0),
					"Grub, Storied Matriarch — return up to one target Goblin card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("up to one target Goblin card from your graveyard", YouOwn(), HasSubtype("Goblin")).WithCount(0, 1)),
			PayToTransformAtYourFirstMain("Grub, Storied Matriarch", "Grub", "{R}"),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Grub, Notorious Auntie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Grants:          []AbilityGrant{SacrificeThisTokenAtEndStepGrant(grubNotoriousAuntieGrant, "Grub, Notorious Auntie — sacrifice this token")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Grub, Notorious Auntie — you may blight 1; if you do, create an attacking copy of the blighted creature",
				func(g *game.Game, item *game.StackItem) error {
					return Blight{
						N:     1,
						May:   true,
						Then:  grubNotoriousAuntieCopy,
						Carry: []uuid.UUID{triggerAttackTarget(item)},
					}.Apply(NewContext(g, item))
				}),
			PayToTransformAtYourFirstMain("Grub, Notorious Auntie", "Grub", "{B}"),
		},
	})
}

// grubNotoriousAuntieGrant is the key of the copy's granted sacrifice
// trigger, an on-disk identity.
const grubNotoriousAuntieGrant = "grub-notorious-auntie/sacrifice-this-token"

// grubNotoriousAuntieCopy is "if you do, create a tapped and attacking
// token that's a copy of the blighted creature, except it has 'At the
// beginning of the end step, sacrifice this token.'" Carry is what Grub
// attacked. An on-disk key.
var grubNotoriousAuntieCopy = RegisterBlighted("blight/grub-notorious-auntie-copy", func(ctx *Context, d BlightDone) error {
	blighted := d.By(ctx.Controller())
	if blighted == uuid.Nil {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, blighted)
	if !ok {
		return nil
	}
	TokenCopyGainsAbility(&tmpl, grubNotoriousAuntieGrant)
	var prefer uuid.UUID
	if len(d.Carry) > 0 {
		prefer = d.Carry[0]
	}
	return CreateTokensAttackingYourChoice{
		Controller: ctx.Controller(),
		Template:   tmpl,
		Copy:       true,
		N:          1,
		Tapped:     true,
		Prefer:     prefer,
		Label:      "Grub, Notorious Auntie",
	}.Apply(ctx)
})
