import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';

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
      <Navbar light expand="md" className="guacNav">
        <NavbarBrand className="guacNavItem" href="/">Avocado</NavbarBrand>
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
            <td>Buy now 🛒</td>
          </tr>
          <tr>
            <th scope="row">2</th>
            <td>Toxicity</td>
            <td>System of a Down</td>
            <td>Toxicity</td>
            <td>0.00243 Eth </td>
            <td>Buy now 🛒</td>
          </tr>
          <tr>
            <th scope="row">3</th>
            <td>BYOB</td>
            <td>System of a Down</td>
            <td>Toxicity</td>
            <td>0.00250 Eth </td>
            <td>Buy now 🛒</td>
          </tr>
          <tr>
            <th scope="row">4</th>
            <td>Sugar</td>
            <td>System of a Down</td>
            <td>System of a Down</td>
            <td>0.00266 Eth </td>
            <td>Buy now 🛒</td>
          </tr>
        </tbody>
      </Table>
    )
  }
}

class App extends Component {
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
