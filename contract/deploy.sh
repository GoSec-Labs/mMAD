#!/bin/bash

# Load environment variables
source .env

echo "🚀 MMAD Token Deployment"
echo "========================"

case "$1" in
    "anvil")
        echo "📡 Deploying to Anvil..."
        eval $FORGE_DEPLOY_ANVIL
        ;;
    "testnet")
        echo "📡 Deploying to BSC Testnet..."
        echo "⚠️  Make sure you have:"
        echo "   - Real private key in BSC_TESTNET_PRIVATE_KEY"
        echo "   - BSC API key in BSC_API_KEY"
        echo "   - Testnet BNB in your wallet"
        echo ""
        read -p "Continue? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            eval $FORGE_DEPLOY_TESTNET
        fi
        ;;
    "testnet-no-verify")
        echo "📡 Deploying to BSC Testnet (without verification)..."
        eval $FORGE_DEPLOY_TESTNET_NO_VERIFY
        ;;
    "build")
        echo "🔨 Building contracts..."
        eval $FORGE_BUILD
        ;;
    "test")
        echo "🧪 Testing contracts..."
        eval $FORGE_TEST
        ;;
    "clean")
        echo "🧹 Cleaning build artifacts..."
        eval $FORGE_CLEAN
        ;;
    *)
        echo "Usage: ./deploy.sh [command]"
        echo ""
        echo "Commands:"
        echo "  anvil              - Deploy to local Anvil"
        echo "  testnet            - Deploy to BSC Testnet (with verification)"
        echo "  testnet-no-verify  - Deploy to BSC Testnet (without verification)"
        echo "  build              - Build contracts"
        echo "  test               - Run tests"
        echo "  clean              - Clean build artifacts"
        echo ""
        echo "Or use the commands directly:"
        echo "  For Anvil: \$FORGE_DEPLOY_ANVIL"
        echo "  For Testnet: \$FORGE_DEPLOY_TESTNET"
        ;;
esac