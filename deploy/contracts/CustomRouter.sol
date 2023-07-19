// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.6.0;

import "https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/utils/math/SafeMath.sol";
import "https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/access/Ownable.sol";
import 'https://github.com/Uniswap/v2-core/blob/master/contracts/UniswapV2Pair.sol';
import "./libraries/TransferHelper.sol";
import "https://github.com/Uniswap/v2-periphery/blob/master/contracts/libraries/UniswapV2Library.sol";

interface IUniSwapV2Factory {
    function INIT_CODE_PAIR_HASH() external view returns(bytes32);
}

contract CustomRouter is Ownable {

    using SafeMath for uint;

    address private factory;
    address private wbnb;

    bytes32 private creationCode;

    constructor(address _factory, address _wbnb) public {
        factory = _factory;
        wbnb = _wbnb;

        creationCode = IUniSwapV2Factory(_factory).INIT_CODE_PAIR_HASH();
    }

    modifier ensure(uint deadline) {
        require(deadline >= block.timestamp, 'Router: EXPIRED');
        _;
    }

    receive() external payable {
        assert(msg.sender == wbnb); // only accept BNB via fallback from the wbnb contract
    }

    function setFactoryAddress(address _factory) external onlyOwner returns(bool success) {
        factory = _factory;
        creationCode = IUniSwapV2Factory(_factory).INIT_CODE_PAIR_HASH(); // creationCode changes too.
        return true;
    }

    function getFactoryAddress() external view onlyOwner returns(address) {
        return factory;
    }

    function setWBNBAddress(address _wbnb) external onlyOwner returns(bool success) {
        wbnb = _wbnb;
        return true;
    }

    function getWBNBAddress() external view onlyOwner returns(address) {
        return wbnb;
    }

    function getCreationCode() external view onlyOwner returns(bytes32) {
        return creationCode;
    }

    function swapExactTokensForTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external virtual ensure(deadline) returns (uint[] memory amounts) {
        amounts = UniSwapV2Library.getAmountsOut(factory, amountIn, path, creationCode);
        require(amounts[amounts.length - 1] >= amountOutMin, 'Router: INSUFFICIENT_OUTPUT_AMOUNT');
        TransferHelper.safeTransferFrom(
            path[0], msg.sender, UniSwapV2Library.pairFor(factory, path[0], path[1], creationCode), amounts[0]
        );

        route(amounts, path, to);
    }

    function route(uint[] memory amounts, address[] memory path, address _to) private {
        for (uint i; i < path.length - 1; i++) {
            (address input, address output) = (path[i], path[i + 1]);
            (address token0,) = UniSwapV2Library.sortTokens(input, output);
            uint amountOut = amounts[i + 1];
            (uint amount0Out, uint amount1Out) = input == token0 ? (uint(0), amountOut) : (amountOut, uint(0));
            address to = i < path.length - 2 ? UniSwapV2Library.pairFor(factory, output, path[i + 2], creationCode) : _to;
            IUniswapV2Pair(UniSwapV2Library.pairFor(factory, input, output, creationCode)).swap(
                amount0Out, amount1Out, to, new bytes(0)
            );
        }
    }
}