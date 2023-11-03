package com.example;

import com.sun.jna.Library;
import com.sun.jna.Native;

public class GoHelloWorld {

  public static interface GoLibrary extends Library {
    void main(String name);
  }

  public static void main(String[] args) {
    GoLibrary library = (GoLibrary) Native.loadLibrary("hello", GoLibrary.class);
    library.main("Bard");
  }
}
