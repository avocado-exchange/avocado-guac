import React, { Component } from 'react';
import './App.css';
import Web3 from 'web3'
import CatalogABI from './contracts/Catalog.json'

import {
  Collapse,
  Navbar,
  NavbarToggler,
  NavbarBrand,
  Nav,
  NavItem,
  NavLink,
  UncontrolledDropdown,
  DropdownToggle,
  DropdownMenu,
  DropdownItem,
  Table
 } from 'reactstrap';

class GuacNav extends Component {
  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.state = {
      isOpen: false
    };
  }
  toggle() {
    this.setState({
      isOpen: !this.state.isOpen
    });
  }
  render() {
    return (
      <Navbar dark expand="md" className="guacNav">
        <NavbarBrand className="guacNavItem" href="/">Avocado <span role="img" aria-label="">🥑</span></NavbarBrand>
        <NavbarToggler onClick={this.toggle} />
        <Collapse isOpen={this.state.isOpen} navbar>
          <Nav className="ml-auto" navbar>
            <NavItem>
              <NavLink href="/">Home</NavLink>
            </NavItem>
            <NavItem>
              <NavLink href="/explore">Explore</NavLink>
            </NavItem>
            <NavItem>
              <NavLink href="/upload">Upload</NavLink>
            </NavItem>
            <UncontrolledDropdown nav inNavbar>
              <DropdownToggle nav caret>
                Options
              </DropdownToggle>
              <DropdownMenu right>
                <DropdownItem>
                  Option 1
                </DropdownItem>
                <DropdownItem>
                  Option 2
                </DropdownItem>
                <DropdownItem divider />
                <DropdownItem>
                  Reset
                </DropdownItem>
              </DropdownMenu>
            </UncontrolledDropdown>
          </Nav>
        </Collapse>
      </Navbar>
    )
  }
}

class SongList extends Component {
  render() {
    return (
      <Table>
        <thead>
          <tr>
            <th>#</th>
            <th>Title</th>
            <th>Artist</th>
            <th>Album</th>
            <th>Price</th>
            <th>Download</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <th scope="row">1</th>
            <td>Chop Suey</td>
            <td>System of a Down</td>
            <td>Toxicity</td>
            <td>0.00250 Eth </td>
            <td>Buy now <span role="img" aria-label="add to cart">🛒</span></td>
          </tr>
          <tr>
            <th scope="row">2</th>
            <td>Toxicity</td>
            <td>System of a Down</td>
            <td>Toxicity</td>
            <td>0.00243 Eth </td>
            <td>Buy now <span role="img" aria-label="add to cart">🛒</span></td>
          </tr>
          <tr>
            <th scope="row">3</th>
            <td>BYOB</td>
            <td>System of a Down</td>
            <td>Toxicity</td>
            <td>0.00250 Eth </td>
            <td>Buy now <span role="img" aria-label="add to cart">🛒</span></td>
          </tr>
          <tr>
            <th scope="row">4</th>
            <td>Sugar</td>
            <td>System of a Down</td>
            <td>System of a Down</td>
            <td>0.00266 Eth </td>
            <td>Buy now <span role="img" aria-label="add to cart">🛒</span></td>
          </tr>
        </tbody>
      </Table>
    )
  }
}

class App extends Component {
  constructor(props) {
    super(props);
    this.web3Provider = new Web3.providers.HttpProvider('http://localhost:9545');
    this.web3 = new Web3(this.web3Provider);
    var web3 = this.web3;
    //const catalogABIParsed = JSON.parse(CatalogABI)

    const myAddr = {from: '0xf17f52151ebef6c7334fad080c5704d77216b732'};

    var catalog = new web3.eth.Contract(CatalogABI.abi,
      '0x8f0483125fcb9aaaefa9209d8e9d7b9c8b9fb90f', {
        from: '0xf17f52151ebef6c7334fad080c5704d77216b732'
      }
    );

    console.log(catalog);
    catalog.methods.nextSongIndexToAssign().call(myAddr)
    .then(console.log)

    /*
    catalog.methods.getListingMetadata(0).call(myAddr)
    .then(console.log)
    */

  }

  render() {
    return (
      <div className="App">
        <GuacNav />
        <br />
        <h3>New Uploads</h3>
        <SongList />
      </div>
    );
  }
}

export default App;
