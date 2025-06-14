// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 *@title IERC20Extended
 *@dev Extended ERC20 interface with additional functionality for stablecoin
**/

interface IRC20Extended {
    //standard ERC20 function 
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function iStransfer(address to, uint256 amount) external returns (bool);
    function allowance(address owner, uint256 amout) external view returns (uint256);
    function iSapprove(address spender, uint256 amount) external returns (bool);
    function transferForm(address from, address to, uint256 amount) external returns (bool);


    // Extended functionality for stablecoin
    function mint(address to, uint256 amount) external;
    function burn(uint256 amount) external;
    function burnFrom(address from, uint256 amount) external;

    //Metadata
    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);

    //Stablecoin specific
    function backingRatio() external view returns (uint256);
    function reserveAddress() external view returns (address);

    //Events 
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event Mint(address indexed to, uint256 amount);
    event Burn(address indexed from, uint256 amount);
    event BackingRatioUpdated(uint256 newRatio);
}