// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

import "./interfaces/IMMadToken.sol";
import "./interfaces/IZKVerifier.sol";
import "./utils/AccessControl.sol";

/**
 * @title MMadToken
 * @dev Zero-Knowledge Moroccan Dirham Stablecoin
 */
contract MMadToken is IMMadToken, AccessControl, ReentrancyGuard, Pausable {
    using Math for uint256;
    using ZKUtils for *;
    using Errors for *;
    using Events for *;
    
    // Token metadata
    string private _name = "Moroccan Mad Stablecoin";
    string private _symbol = "MMAD";
    uint8 private _decimals = 18;
    
    // Token state
    uint256 private _totalSupply;
    uint256 private _maxSupply = 1_000_000_000 * 10**18; // 1 billion MMAD
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    
    // Reserve management
    uint256 private _totalReserves;
    uint256 private _minBackingRatio = 110 * Math.PERCENTAGE_SCALE / 100; // 110%
    address private _reserveManager;
    
    // ZK Integration
    IZKVerifier private _zkVerifier;
    mapping(bytes32 => bool) private _usedProofs;
    uint256 private _lastProofTimestamp;
    
    // Events
    event ReserveManagerSet(address indexed oldManager, address indexed newManager);
    event ZKVerifierSet(address indexed oldVerifier, address indexed newVerifier);
    event MinBackingRatioSet(uint256 oldRatio, uint256 newRatio);
    
    constructor(
        address admin,
        address reserveManager,
        address zkVerifier
    ) {
        if (admin == address(0)) revert Errors.ZeroAddress();
        if (reserveManager == address(0)) revert Errors.ZeroAddress();
        if (zkVerifier == address(0)) revert Errors.ZeroAddress();
        
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setupRole(MINTER_ROLE, admin);
        _setupRole(PAUSER_ROLE, admin);
        _setupRole(RESERVE_MANAGER_ROLE, reserveManager);
        
        _reserveManager = reserveManager;
        _zkVerifier = IZKVerifier(zkVerifier);
        
        emit ReserveManagerSet(address(0), reserveManager);
        emit ZKVerifierSet(address(0), zkVerifier);
    }
    
    // ERC20 Implementation
    function name() public view override returns (string memory) {
        return _name;
    }
    
    function symbol() public view override returns (string memory) {
        return _symbol;
    }
    
    function decimals() public view override returns (uint8) {
        return _decimals;
    }
    
    function totalSupply() public view override returns (uint256) {
        return _totalSupply;
    }
    
    function balanceOf(address account) public view override returns (uint256) {
        return _balances[account];
    }
    
    function transfer(address to, uint256 amount) public override whenNotPaused returns (bool) {
        address owner = msg.sender;
        _transfer(owner, to, amount);
        return true;
    }
    
    function allowance(address owner, address spender) public view override returns (uint256) {
        return _allowances[owner][spender];
    }
    
    function approve(address spender, uint256 amount) public override returns (bool) {
        address owner = msg.sender;
        _approve(owner, spender, amount);
        return true;
    }
    
    function transferFrom(address from, address to, uint256 amount) public override whenNotPaused returns (bool) {
        address spender = msg.sender;
        _spendAllowance(from, spender, amount);
        _transfer(from, to, amount);
        return true;
    }
    
    // Extended Token Functions
    function mint(address to, uint256 amount) public override onlyRole(MINTER_ROLE) whenNotPaused {
        _requireValidMint(amount);
        _mint(to, amount);
    }
    
    function mintWithProof(
        address to,
        uint256 amount,
        IZKVerifier.ProofData calldata proof,
        IZKVerifier.ReserveProof calldata reserveData
    ) external override nonReentrant whenNotPaused {
        _requireValidMint(amount);
        _verifyAndUseMintingProof(proof, reserveData, amount);
        _mint(to, amount);
        
        bytes32 proofHash = ZKUtils.generateProofHash(proof);
        emit Events.TokenMinted(to, amount, proofHash);
    }
    
    function burn(uint256 amount) public override whenNotPaused {
        _burn(msg.sender, amount);
    }
    
    function burnFrom(address from, uint256 amount) public override whenNotPaused {
        _spendAllowance(from, msg.sender, amount);
        _burn(from, amount);
    }
    
    // Reserve Management
    function updateReserves(
        uint256 newReserveAmount,
        IZKVerifier.ProofData calldata proof
    ) external override onlyRole(RESERVE_MANAGER_ROLE) nonReentrant {
        if (!_zkVerifier.verifyReserveProof(proof, IZKVerifier.ReserveProof({
            minRequiredReserve: _totalSupply.safeMul(_minBackingRatio) / Math.PRECISION,
            currentSupply: _totalSupply,
            timestamp: block.timestamp,
            commitment: keccak256(abi.encodePacked(newReserveAmount, block.timestamp))
        }))) {
            revert Errors.InvalidReserveProof();
        }
        
        uint256 oldReserves = _totalReserves;
        _totalReserves = newReserveAmount;
        
        uint256 newRatio = Math.calculateBackingRatio(newReserveAmount, _totalSupply);
        
        emit Events.ReservesUpdated(newReserveAmount, newRatio);
        emit Events.ReserveProofSubmitted(newReserveAmount, _totalSupply, newRatio);
    }
    
    // Compliance Integration
    function transferWithCompliance(
        address to,
        uint256 amount,
        IZKVerifier.ProofData calldata complianceProof,
        bytes32 userHash
    ) external override whenNotPaused returns (bool) {
        if (!_zkVerifier.verifyComplianceProof(complianceProof, userHash, 0)) {
            revert Errors.InvalidComplianceProof();
        }
        
        _transfer(msg.sender, to, amount);
        emit Events.ComplianceProofVerified(userHash, 0);
        
        return true;
    }
    
    // View Functions
    function backingRatio() public view override returns (uint256) {
        return Math.calculateBackingRatio(_totalReserves, _totalSupply);
    }
    
    function reserveAddress() public view override returns (address) {
        return _reserveManager;
    }
    
    function getReserveInfo() external view override returns (
        uint256 totalReserves,
        uint256 requiredReserves,
        uint256 currentBackingRatio,
        uint256 lastProofTimestamp
    ) {
        totalReserves = _totalReserves;
        requiredReserves = _totalSupply.safeMul(_minBackingRatio) / Math.PRECISION;
        currentBackingRatio = backingRatio();
        lastProofTimestamp = _lastProofTimestamp;
    }
    
    function zkVerifier() external view override returns (address) {
        return address(_zkVerifier);
    }
    
    function reserveManager() external view override returns (address) {
        return _reserveManager;
    }
    
    function minBackingRatio() external view override returns (uint256) {
        return _minBackingRatio;
    }
    
    function isPaused() external view override returns (bool) {
        return paused();
    }
    
    function maxSupply() external view override returns (uint256) {
        return _maxSupply;
    }
    
    // Configuration Functions
    function setZKVerifier(address verifier) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (verifier == address(0)) revert Errors.ZeroAddress();
        address oldVerifier = address(_zkVerifier);
        _zkVerifier = IZKVerifier(verifier);
        emit ZKVerifierSet(oldVerifier, verifier);
    }
    
    function setReserveManager(address manager) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (manager == address(0)) revert Errors.ZeroAddress();
        address oldManager = _reserveManager;
        _reserveManager = manager;
        _revokeRole(RESERVE_MANAGER_ROLE, oldManager);
        _grantRole(RESERVE_MANAGER_ROLE, manager);
        emit ReserveManagerSet(oldManager, manager);
    }
    
    function setMinBackingRatio(uint256 ratio) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (ratio < Math.PRECISION) revert Errors.InvalidParameter(); // Must be >= 100%
        uint256 oldRatio = _minBackingRatio;
        _minBackingRatio = ratio;
        emit MinBackingRatioSet(oldRatio, ratio);
    }
    
    // Emergency Functions
    function emergencyWithdraw(address token, uint256 amount) external override onlyRole(DEFAULT_ADMIN_ROLE) whenPaused {
        if (token == address(0)) {
            payable(msg.sender).transfer(amount);
        } else {
            IERC20Extended(token).transfer(msg.sender, amount);
        }
        emit Events.EmergencyWithdrawal(token, amount, msg.sender);
    }
    
    // Internal Functions
    function _transfer(address from, address to, uint256 amount) internal {
        if (from == address(0)) revert Errors.TransferFromZeroAddress();
        if (to == address(0)) revert Errors.TransferToZeroAddress();
        
        uint256 fromBalance = _balances[from];
        if (fromBalance < amount) revert Errors.InsufficientBalance();
        
        unchecked {
            _balances[from] = fromBalance - amount;
            _balances[to] += amount;
        }
        
        emit Transfer(from, to, amount);
    }
    
    function _mint(address to, uint256 amount) internal {
        if (to == address(0)) revert Errors.TransferToZeroAddress();
        if (_totalSupply + amount > _maxSupply) revert Errors.ExceedsMaxSupply();
        
        _totalSupply += amount;
        unchecked {
            _balances[to] += amount;
        }
        
        emit Transfer(address(0), to, amount);
        emit Mint(to, amount);
    }
    
    function _burn(address from, uint256 amount) internal {
        if (from == address(0)) revert Errors.TransferFromZeroAddress();
        
        uint256 accountBalance = _balances[from];
        if (accountBalance < amount) revert Errors.InsufficientBalance();
        
        unchecked {
            _balances[from] = accountBalance - amount;
            _totalSupply -= amount;
        }
        
        emit Transfer(from, address(0), amount);
        emit Burn(from, amount);
    }
    
    function _approve(address owner, address spender, uint256 amount) internal {
        if (owner == address(0)) revert Errors.ApproveFromZeroAddress();
        if (spender == address(0)) revert Errors.ApproveToZeroAddress();
        
        _allowances[owner][spender] = amount;
        emit Approval(owner, spender, amount);
    }
    
    function _spendAllowance(address owner, address spender, uint256 amount) internal {
        uint256 currentAllowance = allowance(owner, spender);
        if (currentAllowance != type(uint256).max) {
            if (currentAllowance < amount) revert Errors.InsufficientAllowance();
            unchecked {
                _approve(owner, spender, currentAllowance - amount);
            }
        }
    }
    
    function _requireValidMint(uint256 amount) internal view {
        if (amount == 0) revert Errors.InvalidAmount();
        if (_totalSupply + amount > _maxSupply) revert Errors.ExceedsMaxSupply();
        
        // Check if reserves support this minting
        uint256 newSupply = _totalSupply + amount;
        if (!Math.meetsMinimumRatio(_totalReserves, newSupply, _minBackingRatio)) {
            revert Errors.InsufficientReserves();
        }
    }
    
    function _verifyAndUseMintingProof(
        IZKVerifier.ProofData calldata proof,
        IZKVerifier.ReserveProof calldata reserveData,
        uint256 mintAmount
    ) internal {
        bytes32 proofHash = ZKUtils.generateProofHash(proof);
        
        if (_usedProofs[proofHash]) revert Errors.ProofAlreadyUsed();
        if (!ZKUtils.validateReserveProof(reserveData)) revert Errors.InvalidReserveProof();
        
        // Verify the proof with the ZK verifier
        if (!_zkVerifier.verifyReserveProof(proof, reserveData)) {
            revert Errors.InvalidReserveProof();
        }
        
        // Mark proof as used
        _usedProofs[proofHash] = true;
        _lastProofTimestamp = block.timestamp;
    }
}