'use strict';

var { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const FabricCAServices = require('fabric-ca-client');
const fs = require('fs');
var SHA256 = require("crypto-js/sha256");
const util = require('util');
const mongoose = require('mongoose')
const channelName = "mychannel"
const chaincodeName = "fabcar"


const getOrg = async (usertype) => {
    let org = null;
    if(usertype === "CSP")
    {
        org = "Org1";
        return org;
    }
    org = "Org2";
    return org;
}

const getCCP = async (org) => {
    let ccpPath = null;
    if(org == 'Org1')
    {
        ccpPath = path.resolve(__dirname, '..', 'config', 'connection-org1.json')
    }
    if(org == 'Org2')
    {
        ccpPath = path.resolve(__dirname, '..', 'config', 'connection-org2.json')
    }
    
    const ccpJSON = fs.readFileSync(ccpPath, 'utf8')
    const ccp = JSON.parse(ccpJSON);
    return ccp
}

const getCaUrl = async (org, ccp) => {
    let caURL = null
    if(org == 'Org1')
    {
        caURL = ccp.certificateAuthorities['ca.org1.example.com'].url
    }
    if(org == 'Org2')
    {
        caURL = ccp.certificateAuthorities['ca.org2.example.com'].url
    }
    return caURL

}

const getWalletPath = async (org) => {
    let walletPath = null
    if(org == 'Org1')
    {
        walletPath = path.join(process.cwd(), 'org1-wallet')
    }
    if(org == 'Org2')
    {
        walletPath = path.join(process.cwd(), 'org2-wallet')
    }
    return walletPath
}

const getAffiliation = async (org) => {
    let aff = null
    if(org == 'Org1')
    {
        aff ='org1.department1'
    }
    if(org == 'Org2')
    {
        aff ='org2.department1'
    }
    return aff
}

const getCaInfo = async (org, ccp) => {
    let caInfo = null
    if(org == 'Org1')
    {
        caInfo = ccp.certificateAuthorities['ca.org1.example.com']
    }
    if(org == 'Org2')
    {
        caInfo = ccp.certificateAuthorities['ca.org2.example.com']
    }
    return caInfo
}


const enrollAdmin = async (org, ccp) => {
    console.log('calling enroll Admin method')
    try {
        const caInfo = await getCaInfo(org, ccp) 
        const caTLSCACerts = caInfo.tlsCACerts.pem;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

        // Create a new file system based wallet for managing identities.
        const walletPath = await getWalletPath(org)
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the admin user.
        const identity = await wallet.get('admin');
        if (identity) {
            console.log('An identity for the admin user "admin" already exists in the wallet');
            return;
        }

        // Enroll the admin user, and import the new identity into the wallet.
        const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
        console.log("Enrollment object is : ", enrollment)
        let x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: `${org}MSP`,
            type: 'X.509',
        };

        await wallet.put('admin', x509Identity);
        console.log('Successfully enrolled admin user "admin" and imported it into the wallet');
        return
    } catch (error) {
        console.error(`Failed to enroll admin user "admin": ${error}`);
    }
}


const isUserRegistered = async (username, userOrg) => {
    try{
        const walletPath = await getWalletPath(userOrg)
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        const userIdentity = await wallet.get(username);
        if (userIdentity) {
            console.log(`An identity for the user ${username} exists in the wallet`);
            return true
        }
        return false
    }
    catch (error) {
        console.error(`Failed to validate": ${error}`);
    }
}

const Register = async (username,usertype) => {
    let userOrg = await getOrg(usertype)
    let ccp = await getCCP(userOrg)
    const caURL = await getCaUrl(userOrg, ccp)
    const ca = new FabricCAServices(caURL);

    const walletPath = await getWalletPath(userOrg)
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);
    const userIdentity = await wallet.get(username);
    if (userIdentity) {
        console.log(`An identity for the user ${username} already exists in the wallet`);
        var response = {
            success: true,
            message: username + ' enrolled Successfully',
        };
        return response
    }

    // Check to see if we've already enrolled the admin user.
    let adminIdentity = await wallet.get('admin');
    if (!adminIdentity) {
        console.log('An identity for the admin user "admin" does not exist in the wallet');
        await enrollAdmin(userOrg, ccp);
        adminIdentity = await wallet.get('admin');
        console.log("Admin Enrolled Successfully")
    }

    // build a user object for authenticating with the CA
    const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, 'admin');
    let secret;
    try {
        // Register the user, enroll the user, and import the new identity into the wallet.
        secret = await ca.register({ affiliation: await getAffiliation(userOrg), enrollmentID: username, role: 'client', attrs: [{ name:"usertype", value: usertype, ecert: true }] }, adminUser);
    } catch (error) {
        return error.message
    }

    const enrollment = await ca.enroll({ enrollmentID: username, enrollmentSecret: secret, attr_reqs: [{ name: "usertype", optional: false }] });
    
    let x509Identity = {
        credentials: {
            certificate: enrollment.certificate,
            privateKey: enrollment.key.toBytes(),
        },
        mspId: `${userOrg}MSP`,
        type: 'X.509',
    };
    await wallet.put(username, x509Identity);
    console.log("Username inserted into the wallet")
    console.log("Registration successful")
    console.log(`Successfully registered and enrolled user ${username} and imported it into the wallet`);

    var response = {
        success: true,
        message: username + ' enrolled Successfully',
    };
    return response
}

module.exports = {
    getCCP: getCCP,
    getWalletPath: getWalletPath,
    isUserRegistered: isUserRegistered,
    Register:Register,
    getOrg:getOrg
}
