/* eslint-disable import/first */
/* jshint ignore:start*/
const electron = window.require("electron");
/* jshint ignore:end*/
import React, { Component } from 'react';
import './App.css';
import Web3 from 'web3'
import CatalogABI from './contracts/Catalog.json'
//import {ipcRenderer} from 'electron';


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
  Table,
  Button,
  ListGroup,
  ListGroupItem,
  Container,
  Row,
  Col
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
            <th>Preview</th>
            <th>Price</th>
            <th>Download</th>
          </tr>
        </thead>
        <tbody>
          {
            this.props.songs.map((song, i) => {
              return <tr key={i}>
                <th scope="row">{i+1}</th>
                <th>{song.title}</th>
                <th>{song.artist}</th>
                <th>{song.album}</th>
                <th>{song.isAvailable ?
                  <Button outline color="info" onClick={
                      () => {this.props.getPreview(song)}
                    }>Preview</Button>
                    : "Unlisted"}</th>
                <td>{song.cost} wei</td>
                <td>
                  <Button outline color="success" onClick={
                      () => {this.props.buy(song)}
                    }>
                    Buy now <span role="img" aria-label="add to cart">🛒</span>
                  </Button>
                </td>
              </tr>
            })
          }
        </tbody>
      </Table>
    )
  }
}

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      listings: [],
      purchasedSongs: []
    };
    this.web3Provider = new Web3.providers.HttpProvider('http://localhost:9545');
    this.web3 = new Web3(this.web3Provider);
    var web3 = this.web3;

    this.catalog = new web3.eth.Contract(CatalogABI.abi,
      '0x8f0483125fcb9aaaefa9209d8e9d7b9c8b9fb90f', {
        from: '0xf17f52151ebef6c7334fad080c5704d77216b732'
      }
    );
  }

  componentDidMount() {
    this.updateListings();
  }

  buy = (song) => {
    /*
    const params = {
      from: '0xf17f52151ebef6c7334fad080c5704d77216b732',
      gas: 3000000,
      value: song.cost
    };
    const catalog = this.catalog;
    catalog.methods.purchaseSong(song.songId).send(params).then(
      (res) => {
        console.log(res);
        var purchasedSongs = this.state.purchasedSongs.append(song.songId);
        this.setState({purchasedSongs});
      }
    )
    */

    electron.ipcRenderer.send("purchase", {
        songId: song.songId,
        title: song.title,
        format: '.mp3',
        account: '0xf17f52151ebef6c7334fad080c5704d77216b732'
    });

  }

  getPreview = (song) => {
    /*electron.ipcRenderer.send("download", {
        url: "http://bekher.me/bekher-cv-ext.pdf",
        properties: {directory: "~/Downloads/"}
    });*/

    electron.ipcRenderer.send("preview", {
        songId: song.songId,
        title: song.title,
        format: '.mp3'
    });
  }

  updateListings = () => {
    const myAddr = {from: '0xf17f52151ebef6c7334fad080c5704d77216b732'};
    const catalog = this.catalog;

    console.log(catalog);
    catalog.methods.nextSongIndexToAssign().call(myAddr)
    .then(lastSongIndex => {
      console.log("last song: "+ lastSongIndex-1);
      var promises = [];
      for (var i = 0; i < lastSongIndex; i++) {
        promises.push(catalog.methods.getListingMetadata(i).call(myAddr))
        promises.push(catalog.methods.getListingInfo(i).call(myAddr))
      }
      return Promise.all(promises);
    }).then(lastSongs => {
      var listings = [];
      for (var i = 0; i < lastSongs.length; i++) {
        const rawListing = lastSongs[i];
        const songInfo = lastSongs[++i];
        const song = {
          songId: i-1,
          filename: this.web3.utils.toAscii(rawListing[0]),
          title: this.web3.utils.toAscii(rawListing[1]),
          album: this.web3.utils.toAscii(rawListing[2]),
          artist: this.web3.utils.toAscii(rawListing[3]),
          genre: this.web3.utils.toAscii(rawListing[4]),
          year: rawListing[5],
          length: rawListing[6],
          seller: songInfo[0],
          cost: songInfo[1],
          isAvailable: songInfo[2],
          previewChunk1Hash: songInfo[3],
          chunk1Key: songInfo[4],
          chunkHashes: songInfo[5]
        }
        listings.push(song);
      }
      listings.reverse();
      console.log(listings);
      this.setState({listings});

    })
  }

  render() {
    return (
      <div className="App">
        <GuacNav />
        <Row>
          <Col />
          <div className="col-md-11">
            <br />
            <h3>Latest listings</h3>
            <SongList songs={this.state.listings} getPreview={this.getPreview} buy={this.buy}/>
            <br />
            <Button outline color="secondary" onClick={this.updateListings}>Update listings</Button>{' '}
            </div>
            <Col />
          </Row>
        </div>
      );
    }
  }

  export default App;
