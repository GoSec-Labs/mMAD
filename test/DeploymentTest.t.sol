// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../src/MMadToken.sol";
import "../src/interfaces/IZKVerifier.sol";
import "./MockZKVerifier.sol"; 

contract DeploymentTest is Test {
    MMadToken public mmadToken;
    MockZKVerifier public zkVerifier; 
    
    address admin = address(0x1);
    
    function setUp() public {
        vm.startPrank(admin);
        
        // Deploy mock ZK verifier
        zkVerifier = new MockZKVerifier();
        
        // Deploy MMad token
        mmadToken = new MMadToken(
            admin,
            admin,
            address(zkVerifier)
        );
        
        vm.stopPrank();
    }
    
    function testDeploymentSuccess() public view {
        assertEq(mmadToken.name(), "Moroccan Mad Stablecoin");
        assertEq(mmadToken.symbol(), "MMAD");
        assertEq(mmadToken.decimals(), 18);
        assertEq(mmadToken.zkVerifier(), address(zkVerifier));
    }
    
    function testMinting() public {
        // Setup reserves with mock proof
        IZKVerifier.ProofData memory mockProof = IZKVerifier.ProofData({
            a: [uint256(1), uint256(2)],
            b: [[uint256(3), uint256(4)], [uint256(5), uint256(6)]],
            c: [uint256(7), uint256(8)],
            publicSignals: new uint256[](1)
        });
        mockProof.publicSignals[0] = 1;
        
        vm.prank(admin);
        mmadToken.updateReserves(1100 * 10**18, mockProof);
        
        vm.prank(admin);
        mmadToken.mint(address(0x2), 1000 * 10**18);
        
        assertEq(mmadToken.balanceOf(address(0x2)), 1000 * 10**18);
        assertEq(mmadToken.totalSupply(), 1000 * 10**18);
    }
    
    function testContractAddresses() public view {
        assertTrue(address(zkVerifier) != address(0));
        assertTrue(address(mmadToken) != address(0));
        
        console.log(" All contracts deployed successfully!");
    }
}