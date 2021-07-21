'use strict';
const log4js = require('log4js');
const logger = log4js.getLogger('BasicNetwork');
const bodyParser = require('body-parser');
const http = require('http')
const util = require('util');
var SHA256 = require("crypto-js/sha256");
const mongoose = require('mongoose')
const express = require('express')
const app = express();
const dbURI = 'mongodb+srv://varun:varun1234@telco-project.rf8w0.mongodb.net/Project?retryWrites=true&w=majority'

const cors = require('cors');
const constants = require('./config/constants.json')

const host = process.env.HOST || constants.host;
const port = process.env.PORT || constants.port;

const helper = require('./app/helper')
const invoke = require('./app/invoke')
const qscc = require('./app/qscc')
const query = require('./app/query')
const PasswordHash = require('./models/schema_pass');
const Customer_Data = require('./models/schema_data');
const Aadhar_Data = require('./models/schema_aadhar');
const { url } = require('inspector');
const CustomerInfo = require('./models/schema_data');
const channelName = "mychannel"
const chaincodeName = "fabcar"

mongoose.connect(dbURI,{useNewUrlParser:true,useUnifiedTopology:true})
    .then((result) => {
        var server = http.createServer(app).listen(port, function () { console.log(`Server started on ${port}`) });
        logger.info('****************** SERVER STARTED AND DATABASE INITIATED ************************');
        logger.info('***************  http://%s:%s  ******************', host, port);
        server.timeout = 240000;
    })
    .catch((err) => console.log(err));


app.use(express.static('public'));
app.use("/css",express.static(__dirname+'public/css'))

app.set('views','./views');
app.set('view engine', 'ejs');

app.options('*', cors());
app.use(cors());
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({
    extended: false
}));

logger.level = 'debug';

function getErrorMessage(field) {
    var response = {
        success: false,
        message: field + ' field is missing or Invalid in the request'
    };
    return response;
}

app.get('/', async function(req,res){
    res.render('index',{title:'Home'})
});

app.get('/CreateCSP',async function (req, res) {
    res.render('register_CSP',{title:"Register"})
});

// Register and enroll user
app.post('/CreateCSP', async function (req, res) {
    var args = {}
    args["name"] = req.body.name;
    args["region"] = req.body.region;
    args["latitude"] = req.body.latitude;
    args["longitude"] = req.body.longitude;
    args["overageRate"] = req.body.overageRate;
    args["roamingRate"] =  req.body.roamingRate;
    args["Doc_type"] = "CSP"
    var password = req.body.password;

    logger.debug('End point : /register');
    logger.debug('Name : ' + args["name"]);
    logger.debug('region  : ' + args["region"]);
    logger.debug('overageRate  : ' + args["overageRate"]);
    logger.debug('roamingRate  : ' + args["roamingRate"]);

    if (!args["name"]) {
        res.json(getErrorMessage('\'name\''));
        return;
    }
    if (!args["region"]) {
        res.json(getErrorMessage('\'region\''));
        return;
    }
    if (!args["overageRate"]) {
        res.json(getErrorMessage('\'overageRate\''));
        return;
    }
    if (!args["roamingRate"]) {
        res.json(getErrorMessage('\'roamingRate\''));
        return;
    }
    
    let response = await helper.Register(args["name"],"CSP");
    console.log(response);
    let resp = await invoke.invokeTransaction("CreateCSP",args["name"],args)
    console.log(resp);
    logger.debug('-- returned from registering the username %s for organization %s', username, orgName);
    if (response && typeof response !== 'string') {
        logger.debug('Successfully registered the username %s for organization %s', username, orgName);
        var pass_hash = SHA256(username+password+"CSP")
        pass_hash = JSON.stringify(pass_hash["words"]);
        console.log(pass_hash);
        const pw_data = new PasswordHash({
            username:username,
            password_hash:pass_hash
        });
        pw_data.save().then((result) => {
            console.log(result);
            res.render('success',{username:username,title:"success"});
        }).catch((err) => {
            console.log(err);
            res.render('failure',{username:username,title:"failed"});
        });
        
    } else {
        logger.debug('Failed to register the username %s for organization %s with::%s', username, orgName, response);
        res.render('failure',{username:username,title:"failed"})
    }
});

// Login 
app.get('/CSPlogin', async function (req, res) {
    res.render('Login',{title:"CSP Login"})
});


app.post('/CSPlogin', async function (req, res) {
    try{
        var username = req.body.name;
        const user_present = await helper.isUserRegistered(username,"Org1")
        console.log(user_present);
        if(!user_present) 
        {
            console.log(`An identity for the user ${username} not exists`);
            var response = {
                success: false,
                message: username + ' was not enrolled',
            };
            return response
        }
        var password = req.body.password;
        var usertype = "CSP";
        var orgName = helper.getOrg(usertype);
        logger.debug('End point : /login');
        logger.debug('User name : ' + username);
        logger.debug('Password  : ' + password);
        if (!username) {
            res.json(getErrorMessage('\'username\''));
            return;
        }
        if (!password) {
            res.json(getErrorMessage('\'Password\''));
            return;
        }
        var pass_hash = SHA256(username+password+usertype)
        PasswordHash.findOne({"username":username},async(err,data)=>{
            if(err)
            {
                res.send(err);
                return;
            }
            else{
                console.log(JSON.stringify(data["password_hash"]));
                console.log(JSON.stringify(pass_hash["words"]));
                if(data["password_hash"] === JSON.stringify(pass_hash["words"]))
                {
                    var url_resp = "/CSPAdmin/"+username;
                    res.redirect(url_resp)
                }
                else{
                    const response_payload = {
                        result: null,
                        error: "Invalid Credentials"
                    }
                    res.send(response_payload)
                }
            }
        });
    }
    catch (error) {
        const response_payload = {
            result: null,
            error: error.name,
            errorData: error.message
        }
        res.send(response_payload)
    }
});

app.get('/CSPAdmin/:username',async function(req,res){
    var username = req.params.username;
    res.render('CSP_admin_page',{title:"CSP Admin",username})
});

app.get('/CSPAdmin/:username/info', async function (req, res) {
    let username = req.params.username;
    let result = await query.query(username,"ReadCSPData",username,"Org1")
    res.send(result);
    // res.render('number_queries',{title:"Get Data",username,result});
});

app.get('/CSPAdmin/:username/GetAllSubscriberSims', async function (req, res) {
    let username = req.params.username;
    let result = await query.query(username,"findAllSubscriberSimsForCSP",username,"Org1")
    res.send(result);
    // res.render('number_queries',{title:"Get Data",username,result});
});

app.get('/CSPAdmin/:username/GetAllSubscriberSims/:publicKey', async function (req, res) {
    let username = req.params.username;
    let publicKey = req.params.publicKey;    
    res.render('display',{title:"Info",username,publicKey})
});

app.get('/CSPAdmin/:username/GetAllSubscriberSims/:publicKey/info', async function (req, res) {
    let username = req.params.username;
    let publicKey = req.params.publicKey; 
    let message = await query.query(publicKey,"ReadSimData",publicKey,"Org2");
    res.render('display_all_services',{title:"Sim Data",message})
});

app.get('/CSPAdmin/:username/GetAllSubscriberSims/:publicKey/history', async function (req, res) {
    let username = req.params.username;
    let publicKey = req.params.publicKey; 
    let message = await query.query(publicKey,"GetHistoryForAsset",publicKey,"Org2");
    res.render('display_all_services',{title:"History Data",message})
});

app.get('/CSPAdmin/:username/GetAllSubscriberSims/:publicKey/calldetails', async function (req, res) {
    let username = req.params.username;
    let publicKey = req.params.publicKey; 
    let message = await query.query(publicKey,"ReadSimData",publicKey,"Org2");
    res.render('display_all_transactions',{title:"Transaction Data",message})
});

app.get('/CSPAdmin/:username/GetAllSubscriberSims/:publicKey/movesim', async function (req, res) {
    let username = req.params.username;
    let publicKey = req.params.publicKey; 
    res.render('display_all_transactions',{title:"Move sim",username,publicKey})
});

app.post('/CSPAdmin/:username/GetAllSubscriberSims/:publicKey/movesim', async function (req, res) {
    try{
        let username = req.params.username;
        let publicKey = req.params.publicKey; 
        let new_loc = req.body.location;
        await invoke.invokeTransaction("MoveSim",publicKey,new_loc)
        console.log("Changing the location is done");
        let operator = await invoke.invokeTransaction("discovery",publicKey);
        console.log("Discovery is done");
        console.log(operator);
        await invoke.invokeTransaction("authentication",publicKey)
        console.log("Authentication is done");
        await invoke.invokeTransaction("UpdateRate",publicKey,operator)
        console.log("UpdateRate is done");
        var url_resp = `/CSPAdmin/${username}/GetAllSubscriberSims/${publicKey}/`
        res.redirect(url_resp)
    }
    catch(error)
    {
        const response_payload = {
            result: null,
            error: error.name,
            errorData: error.message
        }
        res.send(response_payload)
    }
    
});

app.get('/createSubscriberSim',function(req,res){
    res.render('Sim_page',{title:"Dealer Page"})
});


app.post('/createSubscriberSim' ,async function (req,res){
    try{
        var password = req.body.password;
        var args = {};
        args["publicKey"] = req.params.publicKey;
        args["address"] = req.body.address;
        args["msisdn"] = req.body.msisdn;
        args["homeOperatorName"] = req.body.homeOperatorName;
        args["isRoaming"] = "false";
        args["overageThreshold"] = 5;
        args["Doc_type"] = "SubscriberSim"

        let response = await helper.Register(args["publicKey"],"SubscriberSim");
        console.log(response);
        console.log("User created...")

        let message = await invoke.invokeTransaction("CreateSubscriberSim",args["publicKey"],args);
        console.log(message);
        console.log(`message result is : ${message}`)

        await invoke.invokeTransaction("authentication",args["publicKey"])

        var pass_hash = SHA256(args["publicKey"]+password+args["Doc_type"])
        pass_hash = JSON.stringify(pass_hash["words"]);
        console.log(pass_hash);
        const pw_data = new PasswordHash({
            username:username,
            password_hash:pass_hash
        });
        pw_data.save().then((result) => {
            console.log(result);
        }).catch((err) => {
            console.log(err);
        });
            
        res.redirect('/'); 
    }
    catch(error)
    {
        const response_payload = {
            result: null,
            error: error.name,
            errorData: error.message
        }
        res.send(response_payload)
    }
});

app.get('/Userlogin', async function (req, res) {
    res.render('userLogin',{title:"User Login"})
});

app.post('/Userlogin', async function (req, res) {
    var username = req.body.publicKey;
    const user_present = await helper.isUserRegistered(username,"Org2")
    if(!user_present) 
    {
        console.log(`An identity for the user ${username} not exists`);
        var response = {
            success: false,
            message: username + ' was not enrolled',
        };
        return response
    }
    var password = req.body.password;
    var usertype = "SubscriberSim";
    var orgName = helper.getOrg(usertype);
    logger.debug('End point : /login');
    logger.debug('User name : ' + username);
    logger.debug('Password  : ' + password);
    if (!username) {
        res.json(getErrorMessage('\'username\''));
        return;
    }
    if (!password) {
        res.json(getErrorMessage('\'Password\''));
        return;
    }

    var pass_hash = SHA256(username+password+usertype)
    PasswordHash.findOne({"username":username},async (err,data)=>{
        if(err)
        {
            console.log(err);
        }
        else{
            if(data["password_hash"] === JSON.stringify(pass_hash["words"]))
            {
                var url_new = '/user/'+username
                res.redirect(url_new);
            }
            else{
                res.send({success: false, message: "Invalid Credentials" });
            }
        }
    });
});

app.get('/user/:username' ,async function (req,res){
    var publicKey = req.params.username;
    res.render('user_page',{title:"User",number})
});

app.get('/user/:username/info' ,async function (req,res){
    let publicKey = req.params.publicKey; 
    let message = await query.query(publicKey,"ReadSimData",publicKey,"Org2");
    res.send(message);
    // res.render('display_all_services',{title:"Sim Data",message})
});

app.get('/user/:username/calldetails' ,async function (req,res){
    let publicKey = req.params.publicKey; 
    let message = await query.query(publicKey,"ReadSimData",publicKey,"Org2");
    res.send(message)
    // res.render('display_all_transactions',{title:"Transaction Data",message})
});

app.get('/user/:username/simhistory' ,async function (req,res){
    let publicKey = req.params.publicKey; 
    let message = await query.query(publicKey,"GetHistoryForAsset",publicKey,"Org2");
    res.send(message)
    // res.render('display_all_services',{title:"History Data",message})
});

app.get('/user/:username/callout' ,async function (req,res){
    let publicKey = req.params.publicKey; 
    let overageFlag
    let allowOverage

    overageFlag,allowOverage = await invoke.invokeTransaction("VerifyUser",publicKey);

    if(overageFlag === 'false' || (overageFlag === 'true' && allowOverage !== '')) {
        await invoke.invokeTransaction("setOverageFlag",publicKey,allowOverage);
        await invoke.invokeTransaction("callOut",publicKey);
        
    }
    else{
        res.send("Accept the overage charges..")
    }
    res.render('display_all_services',{title:"History Data",message})
});

app.get('/user/:username/callend' ,async function (req,res){
    let publicKey = req.params.publicKey; 

    await invoke.invokeTransaction("callEnd",publicKey)
    await invoke.invokeTransaction("callPay",publicKey)
    res.render('display_all_services',{title:"History Data",message})
});


app.get('/admin/:username/GetIdentity', async function (req, res) {
    try{
        let username = req.params.username
        let message = await query.query(null, "GetSubmittingClientIdentity",username,"Org1");
        const response_payload = {
            result: message,
            error: null,
            errorData: null
        }

        res.send(response_payload);
    }
    catch (error) {
        const response_payload = {
            result: null,
            error: error.name,
            errorData: error.message
        }
        res.send(response_payload)
    }
});





app.get('/channels/:channelName/chaincodes/:chaincodeName', async function (req, res) {
    try {
        logger.debug('==================== QUERY BY CHAINCODE ==================');

        // var channelName = req.params.channelName;
        // var chaincodeName = req.params.chaincodeName;
        console.log(`chaincode name is :${chaincodeName}`)
        let args = req.query.args;
        let fcn = req.query.fcn;

        logger.debug('channelName : ' + channelName);
        logger.debug('chaincodeName : ' + chaincodeName);
        logger.debug('fcn : ' + fcn);
        logger.debug('args : ' + args);

        if (!chaincodeName) {
            res.json(getErrorMessage('\'chaincodeName\''));
            return;
        }
        if (!channelName) {
            res.json(getErrorMessage('\'channelName\''));
            return;
        }
        if (!fcn) {
            res.json(getErrorMessage('\'fcn\''));
            return;
        }
        if (!args) {
            res.json(getErrorMessage('\'args\''));
            return;
        }
        console.log('args==========', args);
        args = args.replace(/'/g, '"');
        args = JSON.parse(args);
        logger.debug(args);

        let message = await query.query(channelName, chaincodeName, args, fcn, req.username, req.orgname);

        const response_payload = {
            result: message,
            error: null,
            errorData: null
        }

        res.send(response_payload);
    } catch (error) {
        const response_payload = {
            result: null,
            error: error.name,
            errorData: error.message
        }
        res.send(response_payload)
    }
});

app.get('/qscc/channels/:channelName/chaincodes/:chaincodeName', async function (req, res) {
    try {
        logger.debug('==================== QUERY BY CHAINCODE ==================');

        var channelName = req.params.channelName;
        var chaincodeName = req.params.chaincodeName;
        console.log(`chaincode name is :${chaincodeName}`)
        let args = req.query.args;
        let fcn = req.query.fcn;
        // let peer = req.query.peer;

        logger.debug('channelName : ' + channelName);
        logger.debug('chaincodeName : ' + chaincodeName);
        logger.debug('fcn : ' + fcn);
        logger.debug('args : ' + args);

        if (!chaincodeName) {
            res.json(getErrorMessage('\'chaincodeName\''));
            return;
        }
        if (!channelName) {
            res.json(getErrorMessage('\'channelName\''));
            return;
        }
        if (!fcn) {
            res.json(getErrorMessage('\'fcn\''));
            return;
        }
        if (!args) {
            res.json(getErrorMessage('\'args\''));
            return;
        }
        console.log('args==========', args);
        args = args.replace(/'/g, '"');
        args = JSON.parse(args);
        logger.debug(args);

        let response_payload = await qscc.qscc(channelName, chaincodeName, args, fcn, req.username, req.orgname);

        // const response_payload = {
        //     result: message,
        //     error: null,
        //     errorData: null
        // }

        res.send(response_payload);
    } catch (error) {
        const response_payload = {
            result: null,
            error: error.name,
            errorData: error.message
        }
        res.send(response_payload)
    }
});