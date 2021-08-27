------------------------------- MODULE hello -------------------------------

EXTENDS Naturals, Sequences, TLC, FiniteSets

(********************
--mpcal hello {
    define {
        HELLO == "hello"
    }

    archetype AHello(ref out) {
    lbl:
        out := HELLO;
    }

    variables out;

    fair process (Hello = 1) == instance AHello(ref out);
}

\* BEGIN PLUSCAL TRANSLATION
--algorithm hello {
  variables out;
  define{
    HELLO == "hello"
  }
  
  fair process (Hello = 1)
  {
    lbl:
      out := HELLO;
      goto Done;
  }
}

\* END PLUSCAL TRANSLATION

********************)
\* BEGIN TRANSLATION (chksum(pcal) = "cb2aaacf" /\ chksum(tla) = "78d542e9")
CONSTANT defaultInitValue
VARIABLES out, pc

(* define statement *)
HELLO == "hello"


vars == << out, pc >>

ProcSet == {1}

Init == (* Global variables *)
        /\ out = defaultInitValue
        /\ pc = [self \in ProcSet |-> "lbl"]

lbl == /\ pc[1] = "lbl"
       /\ out' = HELLO
       /\ pc' = [pc EXCEPT ![1] = "Done"]

Hello == lbl

(* Allow infinite stuttering to prevent deadlock on termination. *)
Terminating == /\ \A self \in ProcSet: pc[self] = "Done"
               /\ UNCHANGED vars

Next == Hello
           \/ Terminating

Spec == /\ Init /\ [][Next]_vars
        /\ WF_vars(Hello)

Termination == <>(\A self \in ProcSet: pc[self] = "Done")

\* END TRANSLATION 

\* Properties

OutOK == <>(out = HELLO)

=============================================================================
\* Modification History
\* Last modified Thu Aug 26 14:12:33 PDT 2021 by shayan
\* Created Thu Aug 26 13:10:19 PDT 2021 by shayan
