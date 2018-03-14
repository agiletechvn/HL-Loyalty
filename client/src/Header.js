import React from "react";
import logo from "./logo.svg";

export default () => (
  <header>
    <img src={logo} className="App-logo" alt="logo" />
    <div id="left_header">Hyperledger Fabric Loyalty Application</div>
    <i id="right_header">Running on Hyperledger Fabric version 1.1.0-preview</i>
  </header>
);
