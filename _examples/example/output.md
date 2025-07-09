🔧 ZK Circuit Examples
======================

📋 Available Circuits:
   📦 Balance Proof v1 (balance_v1)
      Description: Proves that an account balance meets a minimum threshold without revealing the actual balance
      Estimated Time: 5-15 seconds
      Max Constraints: 1000
      Public Inputs: 5
      Private Inputs: 2

   📦 Solvency Proof v1 (solvency_v1)
      Description: Proves that total assets exceed total liabilities by a minimum ratio
      Estimated Time: 30-90 seconds
      Max Constraints: 50000
      Public Inputs: 3
      Private Inputs: 6

💰 Testing Balance Circuit:
   🔍 Testing: User has sufficient balance
   📊 Threshold: 1000
   💎 Balance: 2500 (secret)
   ✅ Test Result: true

🏦 Testing Solvency Circuit:
   🔍 Testing: Institution is solvent with 110% ratio
   💰 Assets: $11 M
   📊 Liabilities: $10 M
   📈 Actual Ratio: 110.0%
   ✅ Test Result: true

⚡ Benchmarking Circuits:
   ⚡ Benchmarking balance_v1:
      🔧 Compile Time: 245ms
      📊 Constraints: 1000
      💾 Memory: ~1 MB
      ⚡ Est. Prove Time: 1-5 seconds
      🔍 Est. Verify Time: 10-50 ms

🧪 Running Test Suites:
   🧪 Running: Balance Circuit Tests
      📝 Test Cases: 3
      ✅ Valid balance above threshold
      🔍 Invalid balance below threshold (expected failure)
      ✅ Edge case: balance equals threshold
      📊 Results: 3/3 passed