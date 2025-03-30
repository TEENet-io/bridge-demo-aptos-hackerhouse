// module my_address::btc_tokenv2 {
//     use std::error;
//     use std::signer;
//     use std::string::{Self, String};
//     use std::vector;
//     use aptos_framework::coin::{Self, BurnCapability, FreezeCapability, MintCapability};
//     use aptos_framework::event;
//     use aptos_framework::account;

//     /// BTC token on Aptos representing cross-chain Bitcoin
//     struct BTC {}

//     /// Storing the mint, burn, freeze capabilities for the BTC token
//     struct BTCCapabilities has key {
//         mint_cap: MintCapability<BTC>,
//         burn_cap: BurnCapability<BTC>,
//         freeze_cap: FreezeCapability<BTC>,
//     }

//     /// Events for tracking bridge operations
//     struct BridgeEvents has key {
//         mint_events: event::EventHandle<MintEvent>,
//         burn_events: event::EventHandle<BurnEvent>,
//     }

//     /// Event emitted when BTC is minted on Aptos
//     struct MintEvent has drop, store {
//         amount: u64,
//         recipient: address,
//         btc_txid: String,  // Bitcoin transaction ID
//     }

//     /// Event emitted when BTC is burned on Aptos for withdrawal to Bitcoin
//     struct BurnEvent has drop, store {
//         amount: u64,
//         burner: address,
//         btc_address: String,  // Bitcoin address for receiving funds
//     }

//     /// Error codes
//     const E_NOT_AUTHORIZED: u64 = 1;
//     const E_ALREADY_INITIALIZED: u64 = 2;
//     const E_INSUFFICIENT_BALANCE: u64 = 3;
//     const E_NOT_IMPLEMENTED: u64 = 4; // Used for functions that aren't fully implemented yet

//     // add admin address
//     fun get_admin_address(): address {
//         @my_address
//     }
    
//     // added: public entry
//     public entry fun initialize_module(account: &signer) {
//         let account_addr = signer::address_of(account);
//         // only admin can initialize
//         assert!(account_addr == get_admin_address(), error::permission_denied(E_NOT_AUTHORIZED));
        
//         // initialize only once
//         assert!(!exists<BTCCapabilities>(account_addr), error::already_exists(E_ALREADY_INITIALIZED));
        
//         let (mint_cap, burn_cap, freeze_cap) = initialize(account);
//         store_capabilities(account, mint_cap, burn_cap, freeze_cap);
//         // Register the admin account to receive BTC
//         register(account);
//     }

//     /// Initialize the BTC token and store the capabilities in the creating account
//     public fun initialize(account: &signer): (MintCapability<BTC>, BurnCapability<BTC>, FreezeCapability<BTC>) {
//         let account_addr = signer::address_of(account);
        
//         // Create the BTC token
//         let (burn_cap, freeze_cap, mint_cap) = coin::initialize<BTC>(
//             account,
//             string::utf8(b"Bitcoin"),
//             string::utf8(b"BTC"),
//             8, // BTC has 8 decimal places
//             true, // Monitor supply
//         );

//         // Register the events
//         if (!exists<BridgeEvents>(account_addr)) {
//             move_to(account, BridgeEvents {
//                 mint_events: account::new_event_handle<MintEvent>(account),
//                 burn_events: account::new_event_handle<BurnEvent>(account),
//             });
//         };
        
//         // Return capabilities in the correct order expected by the storage function
//         (mint_cap, burn_cap, freeze_cap)
//     }

//     /// Save the capabilities in the module creator's account
//     public fun store_capabilities(
//         account: &signer,
//         mint_cap: MintCapability<BTC>,
//         burn_cap: BurnCapability<BTC>,
//         freeze_cap: FreezeCapability<BTC>
//     ) {
//         let account_addr = signer::address_of(account);
        
//         // Ensure capabilities haven't been stored already
//         assert!(!exists<BTCCapabilities>(account_addr), error::already_exists(E_ALREADY_INITIALIZED));
        
//         // Store the capabilities
//         move_to(account, BTCCapabilities {
//             mint_cap,
//             burn_cap,
//             freeze_cap,
//         });
//     }

//     /// Register an account to use BTC
//     public entry fun register(account: &signer) {
//         coin::register<BTC>(account);
//     }

//     /// Mint BTC tokens to a recipient account when BTC is bridged from Bitcoin network
//     public entry fun mint(
//         admin: &signer,
//         recipient: address,
//         amount: u64,
//         btc_txid: String,
//     ) acquires BTCCapabilities, BridgeEvents {
//         let admin_addr = signer::address_of(admin);
        
//         // Verify admin has minting capability
//         assert!(exists<BTCCapabilities>(admin_addr), error::permission_denied(E_NOT_AUTHORIZED));
        
//         // Get minting capability
//         let capabilities = borrow_global<BTCCapabilities>(admin_addr);
        
//         // Make sure recipient is registered
//         if (!coin::is_account_registered<BTC>(recipient)) {
//             // If using a real contract, we'd need proper registration
//             // For now, just abort with proper error
//             abort error::invalid_argument(E_NOT_AUTHORIZED)
//         };
        
//         // Mint tokens
//         let coins = coin::mint<BTC>(amount, &capabilities.mint_cap);
        
//         // Deposit tokens to recipient
//         coin::deposit<BTC>(recipient, coins);
        
//         // Emit mint event
//         let events = borrow_global_mut<BridgeEvents>(admin_addr);
//         event::emit_event(&mut events.mint_events, MintEvent {
//             amount,
//             recipient,
//             btc_txid,
//         });
//     }

//     /// Burn BTC tokens and emit appropriate event, simulating withdrawal to BTC address
//     public fun burn_with_address(
//         account_addr: address, 
//         amount: u64, 
//         btc_receiver_address: vector<u8>
//     ) acquires BTCCapabilities, BridgeEvents {
//         let admin_addr = get_admin_address();
//         let capabilities = borrow_global<BTCCapabilities>(admin_addr);
        
//         let balance = coin::balance<BTC>(account_addr);
//         assert!(balance >= amount, error::invalid_argument(E_INSUFFICIENT_BALANCE));
        
//         let coins_to_burn = coin::mint<BTC>(amount, &capabilities.mint_cap);
//         coin::burn(coins_to_burn, &capabilities.burn_cap);
        
//         let events = borrow_global_mut<BridgeEvents>(admin_addr);
//         event::emit_event(&mut events.burn_events, BurnEvent {
//             amount,
//             burner: account_addr,
//             btc_address: string::utf8(btc_receiver_address),
//         });
//     }

//     /// Standard burn function without receiver address
//     public fun burn(account_addr: address, amount: u64) acquires BTCCapabilities, BridgeEvents {
//         // Call the full version with empty receiver address
//         burn_with_address(account_addr, amount, vector::empty<u8>())
//     }

//     /// Burn tokens from signer's account
//     public entry fun burn_from(account: &signer, amount: u64) acquires BTCCapabilities, BridgeEvents {
//         let account_addr = signer::address_of(account);
//         burn(account_addr, amount)
//     }

//     /// Get the balance of BTC for an account
//     public fun balance(addr: address): u64 {
//         coin::balance<BTC>(addr)
//     }

//     #[test_only]
//     /// Create a test signer capability - this is only for tests
//     public fun create_test_burn_capability(account: &signer): BurnCapability<BTC> acquires BTCCapabilities {
//         let account_addr = signer::address_of(account);
//         let capabilities = borrow_global<BTCCapabilities>(account_addr);
//         coin::create_test_burn_capability(capabilities.burn_cap)
//     }

//     #[test_only]
//     public fun initialize_for_test(account: &signer) {
//         let (mint_cap, burn_cap, freeze_cap) = initialize(account);
//         store_capabilities(account, mint_cap, burn_cap, freeze_cap);
//         // Register the admin account to receive BTC
//         register(account);
//     }
// } 