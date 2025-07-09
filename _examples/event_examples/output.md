🎯 Event System Example
======================

📤 Emitting proof lifecycle events...
   ✅ Proof requested: proof_12345
   ✅ Proof generated: proof_12345 (5s, 1024 bytes)
   ✅ Proof verified: proof_12345 (50ms)

🔧 Emitting circuit events...
   ✅ Circuit compiled: balance_v1 (1000 constraints, 245ms)
   ✅ Circuit compiled: solvency_v1 (15000 constraints, 2s)

❌ Emitting error events...
   ❌ Proof failed: proof_err_1
   ❌ System error: circuit-compiler

🔌 Adding custom handler...
   🔌 Custom handler added
   📈 Custom handler counted: 3 proofs

🚀 Emitting system events...
   🚀 System started: v1.0.0
   🛑 System stopped: graceful shutdown

📊 Event Metrics:
   📊 Total handlers: 4
   📊 Event types: 6
   📊 Queued events: 0
   📈 Event counts