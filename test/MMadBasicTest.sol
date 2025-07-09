// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../src/MMadToken.sol";
import "../src/interfaces/IZKVerifier.sol";
import "./MockZKVerifier.sol"; // ← Use mock instead of real verifier

contract MMadBasicTest is Test {
    MMadToken public mmadToken;
    MockZKVerifier public zkVerifier; // ← Changed to MockZKVerifier
    
    address public admin = makeAddr("admin");
    address public user1 = makeAddr("user1");
    address public user2 = makeAddr("user2");
    address public reserveManager = makeAddr("reserveManager");
    
    function setUp() public {
        vm.startPrank(admin);
        
        // Deploy mock ZK verifier (always returns true)
        zkVerifier = new MockZKVerifier();
        
        // Deploy MMad token with mock verifier
        mmadToken = new MMadToken(
            admin,
            reserveManager,
            address(zkVerifier)
        );
        
        vm.stopPrank();
    }
    
    // Helper function to set up reserves (now works with mock verifier)
    function _setupReserves(uint256 reserveAmount) internal {
        // Create any proof - mock verifier will accept it
        IZKVerifier.ProofData memory mockProof = IZKVerifier.ProofData({
            a: [uint256(1), uint256(2)],
            b: [[uint256(3), uint256(4)], [uint256(5), uint256(6)]],
            c: [uint256(7), uint256(8)],
            publicSignals: new uint256[](1)
        });
        mockProof.publicSignals[0] = 1;
        
        vm.prank(reserveManager);
        mmadToken.updateReserves(reserveAmount, mockProof);
    }
    
    // ==================== KEEP ALL YOUR EXISTING TESTS ====================
    // Just copy-paste all your existing test functions here
    // They'll work now with the mock verifier!
    
    function test_Deployment_Success() public view {
        assertEq(mmadToken.name(), "Moroccan Mad Stablecoin");
        assertEq(mmadToken.symbol(), "MMAD");
        assertEq(mmadToken.decimals(), 18);
        assertEq(mmadToken.totalSupply(), 0);
        assertEq(mmadToken.maxSupply(), 1_000_000_000 * 10**18);
        assertEq(mmadToken.zkVerifier(), address(zkVerifier));
    }
    
    function test_ZKVerifier_Setup() public view {
        assertEq(zkVerifier.getRequiredReserveRatio(), 11000); // 110%
    }
    
    function test_AdminRole_Properly_Set() public view {
        assertTrue(mmadToken.hasRole(mmadToken.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(mmadToken.hasRole(mmadToken.MINTER_ROLE(), admin));
        assertTrue(mmadToken.hasRole(mmadToken.PAUSER_ROLE(), admin));
    }
    
    function test_ReserveManager_Role_Set() public view {
        assertTrue(mmadToken.hasRole(mmadToken.RESERVE_MANAGER_ROLE(), reserveManager));
    }
    
    function test_Unauthorized_Minting_Fails() public {
        vm.prank(user1);
        vm.expectRevert();
        mmadToken.mint(user1, 1000 * 10**18);
    }
    
    function test_Initial_Reserves_Zero() public view {
        (uint256 totalReserves,,,) = mmadToken.getReserveInfo();
        assertEq(totalReserves, 0);
    }
    
    function test_Reserve_Update() public {
        _setupReserves(1000000 * 10**18); // 1M reserves
        
        (uint256 totalReserves,,,) = mmadToken.getReserveInfo();
        assertEq(totalReserves, 1000000 * 10**18);
    }
    
    function test_Basic_Minting_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        assertEq(mmadToken.balanceOf(user1), 1000 * 10**18);
        assertEq(mmadToken.totalSupply(), 1000 * 10**18);
    }
    
    function test_Minting_Fails_Without_Sufficient_Reserves() public {
        _setupReserves(100 * 10**18);
        
        vm.prank(admin);
        vm.expectRevert("Insufficient reserves");
        mmadToken.mint(user1, 1000 * 10**18);
    }
    
    function test_Basic_Transfer_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        mmadToken.transfer(user2, 300 * 10**18);
        
        assertEq(mmadToken.balanceOf(user1), 700 * 10**18);
        assertEq(mmadToken.balanceOf(user2), 300 * 10**18);
    }
    
    function test_Basic_Approval_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        mmadToken.approve(user2, 500 * 10**18);
        
        assertEq(mmadToken.allowance(user1, user2), 500 * 10**18);
    }
    
    function test_Basic_TransferFrom_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        mmadToken.approve(user2, 500 * 10**18);
        
        vm.prank(user2);
        mmadToken.transferFrom(user1, admin, 200 * 10**18);
        
        assertEq(mmadToken.balanceOf(user1), 800 * 10**18);
        assertEq(mmadToken.balanceOf(admin), 200 * 10**18);
        assertEq(mmadToken.allowance(user1, user2), 300 * 10**18);
    }
    
    function test_Basic_Burning_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        mmadToken.burn(300 * 10**18);
        
        assertEq(mmadToken.balanceOf(user1), 700 * 10**18);
        assertEq(mmadToken.totalSupply(), 700 * 10**18);
    }
    
    function test_Max_Supply_Enforced() public {
        vm.prank(admin);
        vm.expectRevert("Exceeds max supply");
        mmadToken.mint(user1, 1_000_000_001 * 10**18);
    }
    
    function test_Mint_At_Max_Supply_With_Huge_Reserves() public {
        _setupReserves(1_100_000_000 * 10**18);
        
        vm.prank(admin);
        mmadToken.mint(user1, 1_000_000_000 * 10**18);
        
        assertEq(mmadToken.totalSupply(), 1_000_000_000 * 10**18);
        assertEq(mmadToken.balanceOf(user1), 1_000_000_000 * 10**18);
    }
    
    function test_Pause_Unpause() public {
        vm.prank(admin);
        mmadToken.pause();
        assertTrue(mmadToken.isPaused());
        
        vm.prank(admin);
        mmadToken.unpause();
        assertFalse(mmadToken.isPaused());
    }
    
    function test_Transfer_Fails_When_Paused() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(admin);
        mmadToken.pause();
        
        vm.prank(user1);
        vm.expectRevert();
        mmadToken.transfer(user2, 100 * 10**18);
    }
    
    function test_Minting_Fails_When_Paused() public {
        _setupReserves(1100 * 10**18);
        
        vm.prank(admin);
        mmadToken.pause();
        
        vm.prank(admin);
        vm.expectRevert();
        mmadToken.mint(user1, 1000 * 10**18);
    }
    
    function test_ZKVerifier_Address_Set() public view {
        assertEq(mmadToken.zkVerifier(), address(zkVerifier));
    }
    
    function test_Reserve_Manager_Update() public {
        address newReserveManager = makeAddr("newReserveManager");
        
        vm.prank(admin);
        mmadToken.setReserveManager(newReserveManager);
        
        assertEq(mmadToken.reserveManager(), newReserveManager);
        assertTrue(mmadToken.hasRole(mmadToken.RESERVE_MANAGER_ROLE(), newReserveManager));
        assertFalse(mmadToken.hasRole(mmadToken.RESERVE_MANAGER_ROLE(), reserveManager));
    }
    
    function test_Backing_Ratio_Update() public {
        vm.prank(admin);
        mmadToken.setMinBackingRatio(120);
        
        assertEq(mmadToken.minBackingRatio(), 120);
    }
    
    function test_Invalid_Backing_Ratio_Fails() public {
        vm.prank(admin);
        vm.expectRevert("Ratio must be >= 100%");
        mmadToken.setMinBackingRatio(90);
    }
    
    function test_Zero_Amount_Transfer_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        mmadToken.transfer(user2, 0);
        
        assertEq(mmadToken.balanceOf(user1), 1000 * 10**18);
        assertEq(mmadToken.balanceOf(user2), 0);
    }
    
    function test_Self_Transfer_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        mmadToken.transfer(user1, 100 * 10**18);
        
        assertEq(mmadToken.balanceOf(user1), 1000 * 10**18);
    }
    
    function test_Burn_More_Than_Balance_Fails() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        vm.expectRevert();
        mmadToken.burn(1001 * 10**18);
    }
    
    function test_Gas_Mint_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        
        vm.prank(admin);
        uint256 gasBefore = gasleft();
        mmadToken.mint(user1, 1000 * 10**18);
        uint256 gasUsed = gasBefore - gasleft();
        
        console.log("Gas used for minting:", gasUsed);
        assertTrue(gasUsed < 100000);
    }
    
    function test_Gas_Transfer_With_Reserves() public {
        _setupReserves(1100 * 10**18);
        vm.prank(admin);
        mmadToken.mint(user1, 1000 * 10**18);
        
        vm.prank(user1);
        uint256 gasBefore = gasleft();
        mmadToken.transfer(user2, 100 * 10**18);
        uint256 gasUsed = gasBefore - gasleft();
        
        console.log("Gas used for transfer:", gasUsed);
        assertTrue(gasUsed < 50000);
    }
}