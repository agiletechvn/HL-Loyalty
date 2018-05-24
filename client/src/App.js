import React, { Component } from "react";
import CustomerList from "./CustomerList";
import Header from "./Header";

import { invoke, query } from "./api";
import "./App.css";

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      customers: [],
      customerHistory: [],
      errorMsg: null,
      txtId: null
    };

    this.customerId = null;
  }

  queryAllCustomers = () => {
    query("get_customers").then(customers => this.setState({ customers }));
  };

  viewCustomerHistory = () => {
    query("get_history", "customer", this.customerId, -1).then(list => {
      const customerHistory = list.map(item => ({
        ...item.Value,
        customerID: item.Value.customerID + "/" + item.TxId
      }));
      this.setState({ customerHistory });
    });
  };

  afterInvoke = data => {
    if (data[0] && data[0].status === "SUCCESS") {
      this.setState({ errorMsg: null, txtId: data[1] });
      this.queryAllCustomers();
      this.viewCustomerHistory();
    } else {
      this.setState({
        errorMsg: "Error: Please enter a valid Customer Id",
        txtId: null
      });
    }
  };

  createCustomer = () => {
    invoke("create_customer", this.customerId).then(data =>
      this.afterInvoke(data)
    );
  };

  changeCustomerName = () => {
    const name = document.querySelector("#customerName").value.trim();
    invoke("update_customer_name", this.customerId, name).then(data =>
      this.afterInvoke(data)
    );
  };

  rewardCustomerCashback = () => {
    const cashback = document.querySelector("#rewardCashback").value.trim();
    invoke("reward_cashback", this.customerId, cashback).then(data =>
      this.afterInvoke(data)
    );
  };

  rewardCustomerToken = () => {
    const token = document.querySelector("#rewardToken").value.trim();
    invoke("reward_token", this.customerId, token).then(data =>
      this.afterInvoke(data)
    );
  };

  render() {
    const { customers, customerHistory, txtId, errorMsg } = this.state;
    return (
      <div className="App">
        <Header />
        <div id="body">
          <div className="row no-gutter">
            <div className="col">
              <h2>Create new Customer</h2>
              <input
                className="form-control m-0"
                placeholder="Ex: 123456789"
                onChange={e => (this.customerId = e.target.value.trim())}
              />
            </div>
            <button
              className="btn btn-primary mb-2"
              onClick={this.createCustomer}
            >
              Create
            </button>
          </div>

          <hr />

          <button
            className="btn btn-primary mb-2"
            onClick={this.queryAllCustomers}
          >
            Query All Customers
          </button>

          <CustomerList list={customers} />

          <hr />
          <div className="row no-gutter">
            <div className="col">
              <h2>Change customer record</h2>
            </div>
            <div className="col">
              Enter an customer id 0000000000:{" "}
              <input
                className="form-control m-0"
                placeholder="Ex: 123456789"
                onChange={e => (this.customerId = e.target.value.trim())}
              />
            </div>
          </div>
          <br />
          {txtId && (
            <h5 className="text-success" id="success_holder">
              Success! Tx ID: {txtId}
            </h5>
          )}
          {errorMsg && (
            <h5 className="text-danger" id="error_holder">
              {errorMsg}
            </h5>
          )}

          <div className="row no-gutter flex">
            <div className="col">
              Enter name of the customer:{" "}
              <input
                className="form-control m-0"
                placeholder="Ex: Barry"
                id="customerName"
              />
            </div>
            <button
              onClick={this.changeCustomerName}
              className="mt-2 btn btn-primary"
            >
              Update
            </button>
          </div>

          <div className="row no-gutter">
            <div className="col">
              Enter amount of cashback reward:{" "}
              <input
                className="form-control m-0"
                placeholder="Ex: 10"
                id="rewardCashback"
              />
            </div>
            <button
              onClick={this.rewardCustomerCashback}
              className="mt-2 btn btn-primary"
            >
              Update
            </button>
          </div>

          <div className="row no-gutter">
            <div className="col">
              Enter customer token reward:{" "}
              <input
                className="form-control m-0"
                placeholder="Ex: 10"
                id="rewardToken"
              />
            </div>
            <button
              onClick={this.rewardCustomerToken}
              className="mt-2 btn btn-primary"
            >
              Update
            </button>
          </div>

          <hr />
          <button
            className="btn btn-primary mb-2"
            onClick={this.viewCustomerHistory}
          >
            View Customer History
          </button>

          <CustomerList list={customerHistory} history={true} />
        </div>
      </div>
    );
  }
}

export default App;
