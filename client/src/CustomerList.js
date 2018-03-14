import React from "react";

export default ({ list, history = false }) => (
  <table className="table" align="center">
    <thead>
      <tr>
        <th>{history ? "ID/TxID" : "ID"}</th>
        <th>Name</th>
        <th>Cashback</th>
        <th>Token</th>
        <th>status</th>
      </tr>
    </thead>
    <tbody>
      {list &&
        list.map(user => (
          <tr key={user.customerID}>
            <td>{user.customerID}</td>
            <td>{user.name}</td>
            <td>{user.cashback}</td>
            <td>{user.token}</td>
            <td>{user.status ? "Active" : "Deactive"}</td>
          </tr>
        ))}
    </tbody>
  </table>
);
