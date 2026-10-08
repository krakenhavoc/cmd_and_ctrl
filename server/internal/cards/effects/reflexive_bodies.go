package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reflexive_bodies.go is the catalog's half of ADR 0041 phase 3, tier 4
// (#1497, Decision P9): every CR 603.12 reflexive trigger a card
// creates names what it does by KEY, registered here with
// game.ReflexiveBody (or game.SimpleDelayedBody / game.DelayedBody for
// an untargeted one), so a table with a reflexive trigger waiting — or
// resolving on the stack — is a restore point.
//
// One file rather than a registration beside each body, for
// delayed_bodies.go's reason: the keys are on-disk identities, in the
// same append-only ledger, and a reviewer adding or renaming one should
// see all of them at once.
//
// A reflexive trigger's own target clause, when it has one, is declared
// HERE, alongside the body — never as a field on the card-side
// ReflexiveTrigger struct — because a reflexive trigger has no catalog
// row for restore to re-derive a captured clause from. The
// game.ReflexiveBody constructor's targetsFrom is handed the trigger's
// source card's instance ID and its Params, both already-carried,
// already-restorable facts:
//
//   - Eden, Seat of the Sanctum's "another target permanent card" reads
//     the instance ID to exclude the source itself.
//   - Teferi Akosa of Zhalfir's mana-value ceiling is X, fixed at
//     creation and carried as Params.Amount.
//   - Every other targeted body here has a clause fixed at
//     registration and ignores both arguments.

var (
	// Generous Plunderer: target opponent creates a tapped Treasure.
	generousPlundererGiftBody = game.ReflexiveBody("generous-plunderer/gift", simpleBody(generousPlundererGift),
		constTargets(func() *game.TargetSpec { return TargetPlayer("target opponent", Opponent()) }))

	// Ziatora, the Incinerator: damage equal to the sacrificed
	// creature's power to any target, plus three Treasures.
	ziatoraFlingBody = game.ReflexiveBody("ziatora/fling", simpleBody(b29ZiatoraFling), constTargets(TargetAny))

	// Breeches, the Blastmaker: "when you win the flip, copy that
	// spell" — untargeted, the spell rides the payload.
	breechesCopyThatSpellBody = game.SimpleDelayedBody("breeches/copy-that-spell", breechesCopyThatSpell)

	// Breeches, the Blastmaker: "when you lose the flip, deal damage
	// equal to that spell's mana value to any target" — the amount is
	// fixed when the flip was lost, so it rides Params.Amount rather
	// than a captured closure.
	breechesBlastBody = game.ReflexiveBody("breeches/blast", breechesBlastEffect, constTargets(TargetAny))

	// Eden, Seat of the Sanctum: "return another target permanent card
	// from your graveyard to your hand". "Another" excludes Eden
	// itself, read off the trigger's own source ID rather than a
	// captured one.
	edenReturnBody = game.ReflexiveBody("eden/return-from-graveyard", simpleBody(edenReturnChosenFromGraveyard),
		func(sourceID uuid.UUID, _ game.EffectParams) *game.TargetSpec {
			return TargetCardInGraveyard("another target permanent card from your graveyard",
				YouOwn(), Permanent(), OtherThan(sourceID))
		})

	// Invasion of Tarkir: "this Siege deals X plus 2 damage to any
	// other target", X the number of Dragon cards revealed.
	tarkirDamageBody = game.ReflexiveBody("tarkir/damage", simpleBody(invasionOfTarkirDamage), constTargets(tarkirAnyOtherTarget))

	// Teferi Akosa of Zhalfir: "shuffle a nonland permanent an opponent
	// controls with mana value X or less into its owner's library", X
	// the number of creatures tapped this way. The body itself needs
	// nothing back — it reads item.Targets — but the clause needs X to
	// build, so it rides Params.Amount.
	teferiAkosaShuffleBody = game.ReflexiveBody("teferi-akosa/shuffle-into-library", simpleBody(teferiAkosaShuffleIntoLibrary),
		func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec { return teferiAkosaShuffleTarget(p.Amount) })

	// Rodolf Duskbringer: "return target creature card with mana value
	// X or less from your graveyard to the battlefield", X the life
	// gained this turn — read live off the game, not off a captured
	// value, so the clause is a constant registration.
	rodolfReturnBody = game.ReflexiveBody("rodolf-duskbringer/return-from-graveyard",
		simpleBody(returnFirstLegalGraveyardTargetToBattlefield),
		constTargets(func() *game.TargetSpec {
			return TargetCardInGraveyard(
				"target creature card in your graveyard with mana value at most the life you gained this turn",
				YouOwn(), Creature(), ManaValueAtMostLifeGainedThisTurn())
		}))

	// Undead Butler: "return target creature card from your graveyard
	// to your hand". The Butler is already in exile by the time this
	// picks, so no "another" clause is needed.
	undeadButlerReturnBody = game.ReflexiveBody("undead-butler/return-to-hand", simpleBody(returnFirstLegalGraveyardTargetToHand),
		constTargets(func() *game.TargetSpec {
			return TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature())
		}))

	// Foray of Orcs: "when you do, this spell deals X damage to target
	// creature an opponent controls, where X is the amassed Army's
	// power". The Army rides the payload and its power is read as the
	// trigger resolves.
	forayOfOrcsDamageBody = game.ReflexiveBody("foray-of-orcs/damage", simpleBody(forayOfOrcsDamage),
		constTargets(func() *game.TargetSpec {
			return TargetCreature("target creature an opponent controls", OpponentControls())
		}))

	// Riddle Gate Gargoyle (ADR 0129 §3): "when you do, target creature
	// you control gains lifelink until end of turn".
	riddleGateLifelinkBody = game.ReflexiveBody("riddle-gate-gargoyle/lifelink",
		simpleBody(firstTargetGainsUntilEndOfTurn(0, "Riddle Gate Gargoyle — lifelink until end of turn", "lifelink")),
		constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))

	// Cyclops Superconductor (ADR 0129 §3): "when you do, this creature
	// deals damage equal to its power to any target". The power is its
	// last-known power as it died, fixed when the energy was paid, and
	// rides Params.Amount; the creature is the damage's source.
	cyclopsSuperconductorDamageBody = game.ReflexiveBody("cyclops-superconductor/damage", breechesBlastEffect, constTargets(TargetAny))

	// Behemoth of Vault 0 (ADR 0129 §3): "when you do, destroy that
	// permanent" — the dies trigger's own target, carried in the
	// payload. Untargeted: nothing new is chosen.
	behemothDestroyBody = game.SimpleDelayedBody("behemoth-of-vault-0/destroy", destroyPayloadPermanents)

	// Territorial Aetherkite (ADR 0129 §3): "when you do, this creature
	// deals that much damage to each other creature". The amount paid
	// rides Params.Amount.
	// Guide of Souls (ADR 0129 §3): "when you do, put two +1/+1 counters
	// and a flying counter on target attacking creature. It becomes an
	// Angel in addition to its other types."
	guideOfSoulsAngelBody = game.ReflexiveBody("guide-of-souls/angel", simpleBody(guideOfSoulsAngel),
		constTargets(func() *game.TargetSpec { return TargetCreature("target attacking creature", AttackingCreature()) }))

	territorialAetherkiteDamageBody = game.DelayedBody("territorial-aetherkite/damage-each-other",
		func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
			return damageEachMatching(NewContext(g, item), And(Creature(), NotSelf(item.SourceCardID)), p.Amount)
		})

	// Blaster Hulk (ADR 0129 §6): "when you do, this creature deals 8
	// damage divided as you choose among up to eight targets". The
	// division is announced with the targets (CR 601.2d).
	blasterHulkDamageBody = game.ReflexiveBody("blaster-hulk/divided-damage", simpleBody(blasterHulkDamage),
		constTargets(func() *game.TargetSpec { return TargetAny().WithCount(0, 8).Dividing(Divide(8)) }))
)

// simpleBody adapts a no-params reflexive body — every one of them
// except Breeches' blast — to game.BodyFunc's three-argument shape, so
// game.ReflexiveBody (which always carries Params, for the bodies that
// need it) has one signature rather than two.
func simpleBody(fn func(g *game.Game, item *game.StackItem) error) game.BodyFunc {
	return func(g *game.Game, item *game.StackItem, _ game.EffectParams) error { return fn(g, item) }
}

// constTargets adapts a target clause that is fixed at registration —
// every reflexive body here except Eden's and Teferi Akosa's — to
// game.ReflexiveBody's targetsFrom shape, which always offers the
// source ID and Params in case a clause needs them.
func constTargets(spec func() *game.TargetSpec) func(uuid.UUID, game.EffectParams) *game.TargetSpec {
	return func(uuid.UUID, game.EffectParams) *game.TargetSpec { return spec() }
}

// breechesBlastEffect is "when you lose the flip, Breeches deals
// damage equal to that spell's mana value to any target". The amount
// is Params.Amount, fixed when the flip was lost.
func breechesBlastEffect(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{
		Source: item.SourceCardID,
		Target: item.Targets[0].ID,
		Amount: p.Amount,
	}.Apply(NewContext(g, item))
}
