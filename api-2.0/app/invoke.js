const { Gateway, Wallets, TxEventHandler, GatewayOptions, DefaultEventHandlerStrategies, TxEventHandlerFactory } = require('fabric-network');
const fs = require('fs');
const EventStrategies = require('fabric-network/lib/impl/event/defaulteventhandlerstrategies');
const path = require("path")
const log4js = require('log4js');
const logger = log4js.getLogger('BasicNetwork');
const util = require('util')

const helper = require('./helper');
const query = require('./query');

const channelName = "mychannel"
const chaincodeName = "fabcar"


const invokeTransaction = async (fcn,username,args) => {
    try {
        let org_name
        if(fcn === "CreateCSP" || fcn === "UpdateCSP" || fcn === "DeleteCSP"){
            org_name = "Org1"
        }
        else {
            org_name = "Org2"
        }
        const ccp = await helper.getCCP(org_name);

        const walletPath = await helper.getWalletPath(org_name);
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        let identity = await wallet.get(username);
        if (!identity) {
            console.log(`An identity for the user ${username} does not exist in the wallet, so registering user`);
            return;
        }

        const connectOptions = {
            wallet, identity: username, discovery: { enabled: true, asLocalhost: true }
        }

        const gateway = new Gateway();
        await gateway.connect(ccp, connectOptions);

        const network = await gateway.getNetwork(channelName);
        const contract = network.getContract(chaincodeName);

        let result;
        let err;
        let message;
        let response;

        switch (fcn) {
            case "UpdateCSP":
            case "UpdateSubscriberSim":
            case "CreateSubscriberSim":
            case "CreateCSP":
                console.log(`User name is ${username}`)
                console.log(JSON.stringify(args));
                result = await contract.submitTransaction('SmartContract:'+fcn, JSON.stringify(args));
                result = {txid: result.toString()}
                break;
        
            case "DeleteCSP":
            case "DeleteSubscriberSim":
            case "Authentication":
            case "CallPay":
                console.log(`User name is ${username}`)
                await contract.submitTransaction('SmartContract:'+fcn,username);
                return;

            case "CallOut":
            case "CallEnd":
                console.log(`User name is ${username}`)
                var time = Math.floor(Date.now()/1000)
                await contract.submitTransaction('SmartContract:'+fcn,username,time);
                return;

            case "CheckForOverage":
                console.log(`User name is ${username}`)
                result = await contract.submitTransaction('SmartContract:'+fcn,username);
                result = result.toString();
                console.log(result);
                return result;

            case "Discovery":
                console.log(`User name is ${username}`)
                let operator = await contract.submitTransaction('SmartContract:'+fcn,username);
                console.log(`Operator found while discovery is ${operator}`);
                return operator;

            case "MoveSim":
            case "UpdateRate":
            case "SetOverageFlag":
                console.log(`User name is ${username}`)
                await contract.submitTransaction('SmartContract:'+fcn,username,args);
                return;

            default:
                break;
        }
        await gateway.disconnect();
        response = {
            message: message,
            result
        }
        return response;
    } catch (error) {
        console.log(`Getting error: ${error}`)
        return error.message
    }
}

exports.invokeTransaction = invokeTransaction;