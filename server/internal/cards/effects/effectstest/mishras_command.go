package effectstest

// mishras_command_test.go — #1112. Mishra's Command is a "choose
// two" with per-mode targets (#764, Kolaghan's Command's shape),
// with an {X} in the cast cost besides: every test here also pins
// that ctx.X() reads back the announced value inside a modal bullet,
// and that a bullet's target is checked against ITS OWN clause (CR
// 608.2b), never the clause of the bullet chosen beside it.

const MishrasCommandOracle = "25434ea3-bcfa-4dae-a16e-10ab87ce32af"
