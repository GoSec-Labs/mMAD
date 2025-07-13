// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

// Import interfaces
import "./interfaces/IMMadToken.sol";
import "./interfaces/IERC20Extended.sol";

// Import libraries
import "./libraries/Math.sol";
import "./libraries/Errors.sol";
import "./libraries/Events.sol";

// Import utils
import "./utils/AccessControl.sol";
import "./utils/ReentrancyGuard.sol";
import "./utils/Pausable.sol";

/**
 * @title MMadToken
 * @dev Zero-Knowledge Moroccan Dirham Stablecoin
 */
contract MMadToken is IMMadToken, IERC20Extended, AccessControl, ReentrancyGuard, Pausable {
    using Math for uint256;
    
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
    uint256 private _minBackingRatio = 110; // 110% (simplified)
    address private _reserveManager;
    
    // ZK Integration
    IZKVerifier private _zkVerifier;
    mapping(bytes32 => bool) private _usedProofs;
    uint256 private _lastProofTimestamp;
    
    constructor(
        address admin,
        address reserveManagerAddr,
        address zkVerifierAddr
    ) {
        require(admin != address(0), "Admin cannot be zero address");
        require(reserveManagerAddr != address(0), "Reserve manager cannot be zero address");
        require(zkVerifierAddr != address(0), "ZK verifier cannot be zero address");
        
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setupRole(MINTER_ROLE, admin);
        _setupRole(PAUSER_ROLE, admin);
        _setupRole(RESERVE_MANAGER_ROLE, reserveManagerAddr);
        
        _reserveManager = reserveManagerAddr;
        _zkVerifier = IZKVerifier(zkVerifierAddr);
        
        emit Events.ReserveManagerUpdated(address(0), reserveManagerAddr);
        emit Events.ZKVerifierUpdated(address(0), zkVerifierAddr);
    }
    
    // ERC20 Implementation
    function name() public view virtual override returns (string memory) {
        return _name;
    }
    
    function symbol() public view virtual override returns (string memory) {
        return _symbol;
    }
    
    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }
    
    function totalSupply() public view virtual override returns (uint256) {
        return _totalSupply;
    }
    
    function balanceOf(address account) public view virtual override returns (uint256) {
        return _balances[account];
    }
    
    function transfer(address to, uint256 amount) public virtual override whenNotPaused returns (bool) {
        address owner = msg.sender;
        _transfer(owner, to, amount);
        return true;
    }
    
    function allowance(address owner, address spender) public view virtual override returns (uint256) {
        return _allowances[owner][spender];
    }
    
    function approve(address spender, uint256 amount) public virtual override whenNotPaused returns (bool) {
        address owner = msg.sender;
        _approve(owner, spender, amount);
        return true;
    }
    
    function transferFrom(address from, address to, uint256 amount) public virtual override whenNotPaused returns (bool) {
        address spender = msg.sender;
        _spendAllowance(from, spender, amount);
        _transfer(from, to, amount);
        return true;
    }
    
    // Extended Token Functions
    function mint(address to, uint256 amount) public virtual override onlyRole(MINTER_ROLE) whenNotPaused {
        _requireValidMint(amount);
        _mint(to, amount);
    }
    
    function mintWithProof(
        address to,
        uint256 amount,
        IZKVerifier.ProofData calldata proof,
        IZKVerifier.ReserveProof calldata reserveData
    ) external virtual override nonReentrant whenNotPaused {
        _requireValidMint(amount);
        _verifyAndUseMintingProof(proof, reserveData, amount);
        _mint(to, amount);
        
        bytes32 proofHash = _generateProofHash(proof);
        emit Events.TokenMinted(to, amount, proofHash);
    }
    
    function burn(uint256 amount) public virtual override whenNotPaused {
        require(amount > 0, "Cannot burn zero amount");
        _burn(msg.sender, amount);
    }
    
    function burnFrom(address from, uint256 amount) public virtual override whenNotPaused {
        require(amount > 0, "Cannot burn zero amount");
        _spendAllowance(from, msg.sender, amount);
        _burn(from, amount);
    }
    
    // Backing ratio and reserve functions
    function backingRatio() public view virtual override returns (uint256) {
        return Math.calculateBackingRatio(_totalReserves, _totalSupply);
    }
    
    function reserveAddress() public view virtual override returns (address) {
        return _reserveManager;
    }
    
    // Reserve Management
    function updateReserves(
        uint256 newReserveAmount,
        IZKVerifier.ProofData calldata proof
    ) external virtual override onlyRole(RESERVE_MANAGER_ROLE) nonReentrant whenNotPaused {

        if (_totalSupply > 0) {
            uint256 requiredMinReserves = (_totalSupply * _minBackingRatio) / 100;
            require(newReserveAmount >= requiredMinReserves, "New reserves insufficient for current     supply backing ratio");
        }

        // Create ReserveProof struct with correct field names
        IZKVerifier.ReserveProof memory reserveProof = IZKVerifier.ReserveProof({
            requiredReserve: (_totalSupply * _minBackingRatio) / 100,
            currentSupply: _totalSupply,
            timestamp: block.timestamp
        });

        
        require(_zkVerifier.verifyReserveProof(proof, reserveProof), "Invalid reserve proof");

        uint256 oldReserves = _totalReserves;
        _totalReserves = newReserveAmount;
        uint256 newRatio = Math.calculateBackingRatio(newReserveAmount, _totalSupply);
        
        emit Events.ReservesUpdated(newReserveAmount, newRatio);
        emit Events.ReserveProofSubmitted(newReserveAmount, _totalSupply, newRatio);
    }

    function _validateReserveAdequacy() internal view {
        if (_totalSupply > 0) {
            uint256 requiredReserves = (_totalSupply * _minBackingRatio) / 100;
            require(_totalReserves >= requiredReserves, "Reserves below minimum backing ratio");
        }
    }
    
    // Compliance Integration
    function transferWithCompliance(
        address to,
        uint256 amount,
        IZKVerifier.ProofData calldata complianceProof,
        bytes32 userHash
    ) external virtual override whenNotPaused returns (bool) {
        require(_zkVerifier.verifyComplianceProof(complianceProof, userHash, 0), "Invalid compliance proof");
        
        _transfer(msg.sender, to, amount);
        emit Events.ComplianceProofVerified(userHash, 0);
        
        return true;
    }
    
    // View Functions from IMMadToken
    function getReserveInfo() external view virtual override returns (
        uint256 totalReserves,
        uint256 requiredReserves,
        uint256 currentBackingRatio,
        uint256 lastProofTimestamp
    ) {
        totalReserves = _totalReserves;
        requiredReserves = (_totalSupply * _minBackingRatio) / 100;
        currentBackingRatio = backingRatio();
        lastProofTimestamp = _lastProofTimestamp;
    }
    
    function zkVerifier() external view virtual override returns (address) {
        return address(_zkVerifier);
    }
    
    function reserveManager() external view virtual override returns (address) {
        return _reserveManager;
    }
    
    function minBackingRatio() external view virtual override returns (uint256) {
        return _minBackingRatio;
    }
    
    function isPaused() external view virtual override returns (bool) {
        return paused();
    }
    
    function maxSupply() external view virtual override returns (uint256) {
        return _maxSupply;
    }
    
    // Configuration Functions
    function setZKVerifier(address verifier) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) {
        require(verifier != address(0), "Verifier cannot be zero address");
        address oldVerifier = address(_zkVerifier);
        _zkVerifier = IZKVerifier(verifier);
        emit Events.ZKVerifierUpdated(oldVerifier, verifier);
    }
    
    function setReserveManager(address manager) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) {
        require(manager != address(0), "Manager cannot be zero address");
        address oldManager = _reserveManager;
        _reserveManager = manager;
        _revokeRole(RESERVE_MANAGER_ROLE, oldManager);
        _grantRole(RESERVE_MANAGER_ROLE, manager);
        emit Events.ReserveManagerUpdated(oldManager, manager);
    }
    


    function setMinBackingRatio(uint256 ratio) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) {
        require(ratio >= 100, "Ratio must be >= 100%");
        require(ratio <= 1000, "Ratio must be <= 1000%");

        if (_totalSupply > 0) {
            uint256 requiredReserves = (_totalSupply * ratio) / 100;
            require(_totalReserves >= requiredReserves, "Current reserves insufficient for new backing ratio");
        }

        uint256 oldRatio = _minBackingRatio;
        _minBackingRatio = ratio;
        emit Events.MinBackingRatioUpdated(oldRatio, ratio);
    }

    function isReserveAdequate() external view returns (bool) {
        if (_totalSupply == 0) return true;
        uint256 requiredReserves = (_totalSupply * _minBackingRatio) / 100;
        return _totalReserves >= requiredReserves;
    }
    
    // Pause functions - Implemented from IMMadToken
    function pause() public virtual override(IMMadToken, Pausable) onlyRole(PAUSER_ROLE) {
        _pause();
        emit Events.EmergencyPause(msg.sender);
    }
    
    function unpause() public virtual override(IMMadToken, Pausable) onlyRole(PAUSER_ROLE) {
        _unpause();
        emit Events.EmergencyUnpause(msg.sender);
    }
    
    // Emergency Functions
    function emergencyWithdraw(address token, uint256 amount) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) whenPaused {
        if (token == address(0)) {
            payable(msg.sender).transfer(amount);
        } else {
            IERC20Extended(token).transfer(msg.sender, amount);
        }
        emit Events.EmergencyWithdrawal(token, amount, msg.sender);
        //SafeERC20.safeTransfer(IERC20(token), msg.sender, amount);
    }
    
    // IZKVerifier functions - implementing interface requirements
    function verifyReserveProof(
        IZKVerifier.ProofData calldata proof,
        IZKVerifier.ReserveProof calldata reserveData
    ) external view virtual override returns (bool) {
        return _zkVerifier.verifyReserveProof(proof, reserveData);
    }
    
    function verifyComplianceProof(
        IZKVerifier.ProofData calldata proof,
        bytes32 userHash,
        uint256 riskScore
    ) external view virtual override returns (bool) {
        return _zkVerifier.verifyComplianceProof(proof, userHash, riskScore);
    }
    
    function verifyBatchProofs(
        IZKVerifier.ProofData[] calldata proofs,
        bytes32[] calldata commitments
    ) external view virtual override returns (bool) {
        return _zkVerifier.verifyBatchProofs(proofs, commitments);
    }
    
    function updateVerificationKey(bytes calldata vkData) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) {
        _zkVerifier.updateVerificationKey(vkData);
    }
    
    function setProofValidator(address validator) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) {
        _zkVerifier.setProofValidator(validator);
    }
    
    function setMaxProofAge(uint256 maxAge) external virtual override onlyRole(DEFAULT_ADMIN_ROLE) {
        _zkVerifier.setMaxProofAge(maxAge);
    }
    
    function getLastProofTimestamp() external view virtual override returns (uint256) {
        return _zkVerifier.getLastProofTimestamp();
    }
    
    function isProofValid(bytes32 proofHash) external view virtual override returns (bool) {
        return _zkVerifier.isProofValid(proofHash);
    }
    
    function getRequiredReserveRatio() external view virtual override returns (uint256) {
        return _zkVerifier.getRequiredReserveRatio();
    }
    
    // Internal Functions
    function _transfer(address from, address to, uint256 amount) internal {
        require(!paused(), "Contract is paused");
        require(from != address(0), "Transfer from zero address");
        require(to != address(0), "Transfer to zero address");

        
        uint256 fromBalance = _balances[from];
        require(fromBalance >= amount, "Insufficient balance");
        
        unchecked {
            _balances[from] = fromBalance - amount;
            _balances[to] += amount;
        }
        
        emit Events.Transfer(from, to, amount);
    }
    
    function _mint(address to, uint256 amount) internal {
        require(!paused(), "Contract is paused");
        require(to != address(0), "Mint to zero address");
        require(_totalSupply + amount <= _maxSupply, "Exceeds max supply");
        
        _totalSupply = _totalSupply + amount;
        _balances[to] = _balances[to] + amount;
        unchecked {
            _balances[to] += amount;
        }
        
        emit Events.Transfer(address(0), to, amount);
        emit Events.Mint(to, amount);
    }
    
    function _burn(address from, uint256 amount) internal {
        require(!paused(), "Contract is paused");
        require(from != address(0), "Burn from zero address");
        require(amount > 0, "Cannot burn zero amount");

        uint256 accountBalance = _balances[from];
        require(accountBalance >= amount, "Insufficient balance");

        // Store pre-burn state for validation
        uint256 totalSupplyBefore = _totalSupply;
        uint256 balanceBefore = accountBalance;

        // Update state atomically
        unchecked {
            _balances[from] = accountBalance - amount;
            _totalSupply -= amount;
        }

        // Comprehensive post-burn state validation
        require(_totalSupply >= 0, "Total supply underflow");
        require(_balances[from] >= 0, "Balance underflow");
        require(_totalSupply == totalSupplyBefore - amount, "Total supply calculation error");
        require(_balances[from] == balanceBefore - amount, "Balance calculation error");

        // Ensure burn doesn't violate economic constraints
        _validateReserveAdequacy();
        
        emit Events.Transfer(from, address(0), amount);
        emit Events.Burn(from, amount);
    }

    function _validateBurnInvariants(address from, uint256 amount) internal view {
        require(_balances[from] >= amount, "Insufficient balance for burn");
        require(_totalSupply >= amount, "Insufficient total supply for burn");

        // Ensure burn won't create invalid state
        uint256 newTotalSupply = _totalSupply - amount;
        uint256 newBalance = _balances[from] - amount;

        require(newTotalSupply >= 0, "Burn would cause supply underflow");
        require(newBalance >= 0, "Burn would cause balance underflow");
    }

    
    function _approve(address owner, address spender, uint256 amount) internal {
        require(!paused(), "Contract is paused");
        require(owner != address(0), "Approve from zero address");
        require(spender != address(0), "Approve to zero address");
        
        _allowances[owner][spender] = amount;
        emit Events.Approval(owner, spender, amount);
    }
    
    function _spendAllowance(address owner, address spender, uint256 amount) internal {
        uint256 currentAllowance = allowance(owner, spender);
        if (currentAllowance != type(uint256).max) {
            require(currentAllowance >= amount, "Insufficient allowance");
            unchecked {
                _approve(owner, spender, currentAllowance - amount);
            }
        }
    }
    
    function _requireValidMint(uint256 amount) internal view {
        require(!paused(), "Contract is paused");
        require(amount > 0, "Invalid amount");
        require(_totalSupply + amount <= _maxSupply, "Exceeds max supply");

        // Check if reserves support this minting
        uint256 newSupply = _totalSupply + amount;
        uint256 requiredReserves = (newSupply * _minBackingRatio) / 100;
        require(_totalReserves >= requiredReserves, "Insufficient reserves for backing ratio");

        require(newSupply > 0, "Invalid supply calculation");
    }
    
    function _verifyAndUseMintingProof(
        IZKVerifier.ProofData calldata proof,
        IZKVerifier.ReserveProof calldata reserveData,
        uint256 mintAmount
    ) internal {
        bytes32 proofHash = _generateProofHash(proof);
        
        require(!_usedProofs[proofHash], "Proof already used");
        require(_validateReserveProof(reserveData), "Invalid reserve proof");
        
        // Verify the proof with the ZK verifier
        require(_zkVerifier.verifyReserveProof(proof, reserveData), "ZK verification failed");
        
        // Mark proof as used
        _usedProofs[proofHash] = true;
        _lastProofTimestamp = block.timestamp;
    }
    
    function _validateReserveProof(IZKVerifier.ReserveProof calldata reserveData) internal view returns (bool) {
        return reserveData.requiredReserve > 0 && 
               reserveData.currentSupply > 0 && 
               reserveData.timestamp <= block.timestamp &&
               reserveData.timestamp > block.timestamp - 1 hours;
    }
    
    function _generateProofHash(IZKVerifier.ProofData calldata proof) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(proof.a, proof.b, proof.c, proof.publicSignals));
    }
}