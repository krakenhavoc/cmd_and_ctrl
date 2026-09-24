package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battles.go — the card-facing half of S27's battle support
// (CR 310). The lifecycle is engine-side and keys off the card type:
// defense counters on entry, the protector prompt, and the CR 704.5v/w
// sweep at zero defense are all in server/internal/game/battle.go and
// apply to a battle the catalog has never heard of.
//
// What a card file declares here is:
//
//	BattleSpec     the printed defense, as a FALLBACK for cards that
//	               never go through deck import
//	DefeatedTrigger the "when this Siege is defeated" ability
//
// A battle's own ETB trigger is an ordinary EventETB TriggeredAbility
// and needs nothing special.

// BattleSpec is a battle's printed battle data — the catalog's
// fallback copy of what the deck importer stamps onto
// game.Card.StartingDefense from Scryfall.
//
// LEAVE IT ALONE FOR A REAL CARD. Defense is printed card data like
// power, toughness and starting loyalty, not card-effect data: the
// importer parses Scryfall's per-face `defense` and the engine stamps
// the counters on every battlefield entry, catalog entry or not. This
// slot exists for the cards that never go through import — tokens,
// fixtures, the demo seed — exactly as Spec.StartingLoyalty does
// after #274. Setting it on a card a player can actually own is
// redundant at best and a second source of truth at worst.
//
// It is declared, and the Sieges in this batch do set it, because a
// battle is a new enough card type that a reader deserves to see the
// number next to the rules text — and because that makes the fallback
// path itself testable without a Scryfall fixture.
type BattleSpec struct {
	// Defense is the printed defense (CR 310.4) — the number of
	// defense counters the battle enters with.
	Defense int

	// Subtype is the battle's printed subtype: "Siege" today, and
	// nothing else has been printed. Carried so a future
	// "battles you control" or per-subtype rule has something to
	// read that is not a substring of the type line.
	Subtype string
}

// BattleSubtypeSiege is the only printed battle subtype (CR 310.12).
// A Siege is the one that chooses a protector and that transforms
// when defeated.
const BattleSubtypeSiege = "Siege"

// DefeatedTrigger declares a battle's "when this is defeated"
// ability (CR 310.12b) — the last defense counter has come off and the
// battle is about to leave the battlefield.
//
// It is an ordinary triggered ability on an ordinary event, so it
// uses the stack and can be responded to. What is NOT ordinary is
// the timing: the engine announces the defeat from the state-based
// action pass, immediately BEFORE the CR 704.5v/w move puts the battle
// in the graveyard. So Build runs with the battle still on the
// battlefield and the Effect runs after it has left — which is why
// the effect must read the battle by id off the item rather than
// assuming where it is.
func DefeatedTrigger(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return On(game.EventBattleDefeated, Self, label, effect)
}

// SiegeBackFace is the face index every printed Siege's back half
// sits on. Scryfall puts the battle first on every one of the 36
// `transform` battles in the dump, so the back is always face 1 —
// named rather than written as a literal because it is the value the
// grant, the catalog key ("<oracle_id>#1") and faceOnResolve all have
// to agree on.
const SiegeBackFace = 1

// SiegeTransformedCastCaveat is the one player-facing sentence every
// Siege in the catalog owes its reader, because the timing of the
// transformed cast is the single way the engine's Sieges differ from
// the printed ones. Shared rather than retyped per card: it is one
// mechanism (see SiegeDefeated), and three copies of a sentence drift
// the moment the mechanism changes.
const SiegeTransformedCastCaveat = "When the Siege is defeated you get to cast its back face for free, but only during your own main phase and only before the turn ends — so a Siege defeated on someone else's turn is lost."

// SiegeDefeated is the reminder-text half every printed Siege shares:
// "exile it, then cast it transformed" (CR 310.12b).
//
// The battle leaves the battlefield for exile, and the exiled card is
// stamped with a grant that opens its BACK face, for nothing, to the
// player who controlled the battle. That is the S32 half of the seam
// S27 named: CastPermission.Faces says which faces a per-instance
// grant opens, CastSpell's exile branch reads it before the face gate
// (faceForCastLocked), and faceOnResolve keeps a cast `transform`
// face rather than forcing every permanent front-up. Without all
// three the cast would have announced Refraction Elemental and
// resolved into Invasion of Karsus — a defeated battle re-entering
// the battlefield, which is worse than not casting it at all.
//
// # The price
//
// "{0}", not "" — CastPermission.Cost reads the empty
// string as "no override, pay the printed cost", and a Siege back
// face HAS no printed cost (Scryfall gives every one of them
// `mana_cost: ""`), so the printed path would charge nothing by
// accident rather than by instruction. Spelling it {0} means the
// parser never sees the blank cost at all. Same reasoning, same
// idiom, as cascade's grantFreeCastLocked.
//
// # SANDBOX SIMPLIFICATION, weaker than printed: a grant, not an
// inline cast
//
// Printed, the cast happens as the defeated trigger resolves and
// ignores timing. Here the trigger stamps a permission and the
// controller casts it with an ordinary cast_spell action, so the back
// face obeys sorcery speed and the grant lapses at this turn's
// cleanup. The common line still works — you attack your own Siege,
// it is defeated in the combat damage step, and you cast the back
// face in your postcombat main phase — but a Siege defeated on
// someone else's turn is lost, because its controller has no sorcery
// window left before the grant expires.
//
// That is exactly the trade cascade already made and for the same
// reason (see cascade.go): an inline cast would have to collect
// targets, modes and X from inside a resolution, and the announce
// path has no frame for a half-validated cast. It is the weaker
// direction in every case — a cast that may not happen, never one
// that happens twice or later than printed allows. The grant is
// bounded to THIS turn precisely so that it cannot become the
// stronger thing: a window the player may keep open until a better
// moment is not what the card grants.
//
// A back face left uncast stays in exile, which is where printed
// leaves a Siege whose transformed cast is countered, so nothing has
// to sweep it.
func SiegeDefeated() func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		battleID := item.SourceCardID
		if battleID == uuid.Nil {
			return nil
		}
		// A Siege with no imported back face — a fixture, a token, a
		// demo-seed row — gets the exile and no grant. Granting one
		// would be actively wrong rather than merely incomplete:
		// SetFace clamps an out-of-range index to 0, so a grant for a
		// face that does not exist would offer the BATTLE itself as a
		// free cast out of exile. Exiling is the half that is right
		// either way (#259).
		if c, ok := g.LookupCardForEffect(battleID); !ok || c.FaceCount() <= SiegeBackFace {
			return g.ExileCardForEffect(battleID)
		}
		// The battle is already in its owner's graveyard by the time
		// this resolves: the defeat is announced from the SBA pass,
		// and the CR 704.5v/w move runs in the same pass, before the
		// trigger drains onto the stack. The exile helper finds it
		// wherever it is, and drops the grant silently if a trigger
		// off the move takes the card somewhere else first.
		//
		// The grant names the ability's controller, which is the
		// battle's controller — the player who cast it, not the one
		// who defeated it. A Siege is protected by an opponent and
		// attackable by everyone else, so the two are routinely
		// different players and the card is unambiguous about which
		// one gets the back face.
		return g.ExileCardWithPermissionForEffect(battleID, game.CastPermission{
			Player:   item.Controller,
			Cost:     "{0}",
			CastOnly: true,
			Faces:    []int{SiegeBackFace},
			Label:    "Cast it transformed, without paying its mana cost",
		})
	}
}
