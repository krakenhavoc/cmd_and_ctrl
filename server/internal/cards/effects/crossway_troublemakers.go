package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crossway Troublemakers — Creature — Vampire {5}{B}, 5/5 (EDHREC
// rank 3630):
//
//	"Attacking Vampires you control have deathtouch and lifelink.
//	 (Any amount of damage they deal to a creature is enough to
//	 destroy it. Damage dealt by those creatures also causes their
//	 controller to gain that much life.)
//	 Whenever a Vampire you control dies, you may pay 2 life. If you
//	 do, draw a card."
//
// The Vampire tribal six-drop. The printed static is a layer 6 grant
// gated on Card.AttackingTarget (AttackingCreaturesHave, declared
// DependsOnAttackingStatus so the cache is rebuilt as attackers are
// declared, put onto the battlefield attacking, or leave combat), on
// every Vampire the controller controls — the Troublemakers
// themselves included, and a Vampire that enters attacking too. The
// dies trigger
// fires for any Vampire creature the controller controlled — a
// Vampire as it last existed on the battlefield (CR 603.10a, #1679),
// grants included — the Troublemakers' own death included; its "you may pay 2 life" is
// the ordinary trigger prompt, and a controller who cannot pay when
// it resolves neither pays nor draws.
//
// No simplification.
//
// The pay-2-life decision used to be taken when the trigger fired
// (the CR 603.5 "you may" prompt the harvester asks before the
// ability goes on the stack), because that was the only yes/no the
// engine had. #796 gave a resolving effect its own, so the question
// is now asked where the card prints it — as the ability RESOLVES,
// after the response window, which is when a player actually knows
// whether the 2 life is affordable.
func init() {
	Register(Spec{
		OracleID:     "1a362e4d-6c02-4b67-ab63-c6622e505195",
		Name:         "Crossway Troublemakers",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			crosswayTroublemakersVampires(),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b34VampireYouControlDied(ev, source, g)
			}, "Crossway Troublemakers — pay 2 life, draw a card", crosswayMayPayTwoLifeToDraw),
		},
	})
}

// crosswayTroublemakersVampires is "Attacking Vampires you control have
// deathtouch and lifelink". "Vampires" is a word of the source's text
// (CR 612.2, #3984), so the shared static's scope is narrowed here,
// where the static's source is read.
func crosswayTroublemakersVampires() game.StaticAbility {
	s := AttackingCreaturesHave("Attacking Vampires you control have deathtouch and lifelink.", nil, "deathtouch", "lifelink")
	base := s.AppliesTo
	s.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return base(target, g, source) && target.HasSubtype(source.TextSubtype("Vampire"))
	}
	return s
}

// crosswayMayPayTwoLifeToDraw is the printed body: "you may pay 2
// life. If you do, draw a card."
//
// LifeCost is declared so the move list prices it (#547). Without it
// a bot at 2 life answers "pay" and dies — the branch re-checks
// affordability anyway (life moves between the question and the
// answer), but a policy that cannot see the price never gets that far.
//
// Caller holds g.mu.
func crosswayMayPayTwoLifeToDraw(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Crossway Troublemakers — a Vampire died. Pay 2 life to draw a card?",
		YesLabel: "Pay 2 life",
		NoLabel:  "Decline",
		LifeCost: crosswayLifePayment,
		OnYes:    crosswayPayTwoLifeThenDraw,
	}.Apply(NewContext(g, item))
}

// crosswayPayTwoLifeThenDraw is the "if you do" branch: a CR 118.3
// cost, then the linked draw. The body is the shared one in
// helpers.go — Erebos, Bleak-Hearted prints the same sentence.
//
// Caller holds g.mu.
func crosswayPayTwoLifeThenDraw(ctx *Context) error {
	return payLifeThenDraw(ctx, crosswayLifePayment, 1)
}

// crosswayLifePayment is the printed 2.
const crosswayLifePayment = 2
