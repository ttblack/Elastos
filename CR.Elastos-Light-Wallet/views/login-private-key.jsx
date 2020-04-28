const React = require('react');

module.exports = (props) => {
  const App = props.App;
  const GuiToggles = props.GuiToggles;
  const usePrivateKey = () => {
    const success = App.getPublicKeyFromPrivateKey();
    if(success) {
      GuiToggles.showHome();
    }
  }
  return (
  <div id="loginPrivateKey">
  <div className="login-div ">
    <div className="flex_center w100pct">
    <img className="flex1" src="artwork/voting-back.svg" height="26px" width="26px" onClick={(e)=> GuiToggles.showLanding()}/>
    <img src="artwork/logonew.svg" height="41px" width="123px" />
    <div className="flex1"></div>
    </div>
    <p className="address-text font_size24 margin_none display_inline_block gradient-font">Enter Private Key</p>
    <textarea className="qraddress-div color_white textarea-placeholder padding_5px" type="text" rows="4" cols="50" id="privateKey" placeholder="Enter Private Key"></textarea>
    <div className="flex_center">
    <button className="proceed-btn dark-hover" onClick={(e)=> usePrivateKey()}>
          <p>Proceed</p>
          </button>
  </div>

  <ul className="color_white list-none padding_0">
    <li className="gradient-font font_size14" >Tips</li>
    <li>Enter your Private Key above.</li>
    <li>Your Private Key is a string of numbers and letters.</li>
    <li>Please use the Mnemonic login if you have your 12 seed words.</li>
    <li>Please take precautions when entering your Private Key, make sure nobody is watching you physically or virtually.</li>
  </ul>
  </div>
</div>);
}














//   (
//   <table id="loginPrivateKey" className="bordered w750h520px">
//     <tbody>
//       <tr>
//         <td colSpan="2">
//           <div>Private Key</div>
//         </td>
//       </tr>
//       <tr>
//         <td colSpan="2">
//           <input className="monospace" type="text" size="64" id="privateKey" placeholder="Private Key"></input>
//         </td>
//       </tr>
//       <tr>
//         <td className="ta_left">
//           <div className="bordered bgcolor_black_hover display_inline_block" onClick={(e)=> usePrivateKey()}>Use Private Key</div>
//         </td>
//         <td className="ta_right">
//           <div className="bordered bgcolor_black_hover display_inline_block ta_right" onClick={(e)=> GuiToggles.showLanding()}>Back</div>
//         </td>
//       </tr>
//     </tbody>
//   </table>
//   );
// }
