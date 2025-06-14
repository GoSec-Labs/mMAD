// SPDX-License-Identifier: SEE LICENSE IN LICENSE
pragma solidity ^0.8.19;

/**
 * @title AccessControl
 * @dev Role-based access control mechanism
 */
abstract contract AccessControl {
    using Errors for *;
    using Events for *;
    
    struct RoleData {
        mapping(address => bool) members;
        bytes32 adminRole;
    }
    
    mapping(bytes32 => RoleData) private _roles;
    
    bytes32 public constant DEFAULT_ADMIN_ROLE = 0x00;
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    bytes32 public constant RESERVE_MANAGER_ROLE = keccak256("RESERVE_MANAGER_ROLE");
    bytes32 public constant PROOF_VALIDATOR_ROLE = keccak256("PROOF_VALIDATOR_ROLE");
    
    modifier onlyRole(bytes32 role) {
        _checkRole(role);
        _;
    }
    
    function hasRole(bytes32 role, address account) public view returns (bool) {
        return _roles[role].members[account];
    }
    
    function getRoleAdmin(bytes32 role) public view returns (bytes32) {
        return _roles[role].adminRole;
    }
    
    function grantRole(bytes32 role, address account) public onlyRole(getRoleAdmin(role)) {
        _grantRole(role, account);
    }
    
    function revokeRole(bytes32 role, address account) public onlyRole(getRoleAdmin(role)) {
        _revokeRole(role, account);
    }
    
    function renounceRole(bytes32 role, address account) public {
        if (account != msg.sender) revert Errors.Unauthorized();
        _revokeRole(role, account);
    }
    
    function _setupRole(bytes32 role, address account) internal {
        _grantRole(role, account);
    }
    
    function _setRoleAdmin(bytes32 role, bytes32 adminRole) internal {
        bytes32 previousAdminRole = getRoleAdmin(role);
        _roles[role].adminRole = adminRole;
        emit Events.RoleAdminChanged(role, previousAdminRole, adminRole);
    }
    
    function _grantRole(bytes32 role, address account) internal {
        if (!hasRole(role, account)) {
            _roles[role].members[account] = true;
            emit Events.RoleGranted(role, account, msg.sender);
        }
    }
    
    function _revokeRole(bytes32 role, address account) internal {
        if (hasRole(role, account)) {
            _roles[role].members[account] = false;
            emit Events.RoleRevoked(role, account, msg.sender);
        }
    }
    
    function _checkRole(bytes32 role) internal view {
        if (!hasRole(role, msg.sender)) revert Errors.Unauthorized();
    }
}